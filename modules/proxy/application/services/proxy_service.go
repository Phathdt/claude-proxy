package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"
	"claude-proxy/modules/proxy/infrastructure/clients"

	sctx "github.com/phathdt/service-context"
)

// ProxyService implements the proxy business logic
type ProxyService struct {
	accountRepo  interfaces.AccountRepository
	accountSvc   interfaces.AccountService
	claudeClient *clients.ClaudeAPIClient
	logger       sctx.Logger
}

// NewProxyService creates a new proxy service
func NewProxyService(
	accountRepo interfaces.AccountRepository,
	accountSvc interfaces.AccountService,
	claudeClient *clients.ClaudeAPIClient,
	logger sctx.Logger,
) interfaces.ProxyService {
	return &ProxyService{
		accountRepo:  accountRepo,
		accountSvc:   accountSvc,
		claudeClient: claudeClient,
		logger:       logger,
	}
}

// ProxyRequest proxies an HTTP request to Claude API
func (s *ProxyService) ProxyRequest(
	ctx context.Context,
	token *entities.Token,
	req *http.Request,
) (*http.Response, error) {
	// Get valid account
	account, err := s.GetValidAccount(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Withs(sctx.Fields{
		"token_id":     token.ID,
		"token_name":   token.Name,
		"account_id":   account.ID,
		"account_name": account.Name,
		"org_uuid":     account.OrganizationUUID,
		"method":       req.Method,
		"path":         req.URL.Path,
	}).Info("Proxying request to Claude API")

	// Get valid access token (will refresh if needed)
	accessToken, err := s.accountSvc.GetValidToken(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid access token: %w", err)
	}

	// Read request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	// Validate and fix extended thinking parameters if needed
	if len(bodyBytes) > 0 {
		bodyBytes, err = s.validateAndFixThinkingParams(bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to validate request parameters: %w", err)
		}
	}

	// Build path with query string
	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}

	// Proxy the request - only pass access token and body, headers are built in claude_client
	resp, err := s.claudeClient.ProxyRequest(ctx, req.Method, path, accessToken, bodyBytes)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"error":      err.Error(),
			"token_id":   token.ID,
			"account_id": account.ID,
		}).Error("Failed to proxy request")
		return nil, fmt.Errorf("failed to proxy request: %w", err)
	}

	s.logger.Withs(sctx.Fields{
		"status_code": resp.StatusCode,
		"token_id":    token.ID,
		"account_id":  account.ID,
	}).Info("Received response from Claude API")

	return resp, nil
}

// GetValidAccount returns a valid active account using load balancing strategy
// Priority: accounts that don't need token refresh, then rotate through all active accounts
func (s *ProxyService) GetValidAccount(ctx context.Context) (*entities.Account, error) {
	accounts, err := s.accountRepo.GetActiveAccounts(ctx)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no active accounts available")
	}

	// Filter accounts that don't need token refresh (healthier accounts)
	var healthyAccounts []*entities.Account
	for _, acc := range accounts {
		if !acc.NeedsRefresh() {
			healthyAccounts = append(healthyAccounts, acc)
		}
	}

	// If we have healthy accounts, select from them with round-robin
	var selectedAccounts []*entities.Account
	if len(healthyAccounts) > 0 {
		selectedAccounts = healthyAccounts
	} else {
		// Fallback to all active accounts if none are healthy
		selectedAccounts = accounts
	}

	// Round-robin selection using account ID hash to distribute load
	// This ensures different requests rotate through accounts
	account := s.selectAccountRoundRobin(selectedAccounts)

	s.logger.Withs(sctx.Fields{
		"account_id":       account.ID,
		"account_name":     account.Name,
		"needs_refresh":    account.NeedsRefresh(),
		"total_accounts":   len(accounts),
		"healthy_accounts": len(healthyAccounts),
	}).Debug("Selected account for proxy request")

	return account, nil
}

// selectAccountRoundRobin selects an account using round-robin strategy
// Uses a simple hash-based distribution to avoid needing persistent state
func (s *ProxyService) selectAccountRoundRobin(accounts []*entities.Account) *entities.Account {
	if len(accounts) == 0 {
		return nil
	}
	if len(accounts) == 1 {
		return accounts[0]
	}

	// Use current timestamp as a rotating index
	// This provides round-robin behavior without needing to maintain state
	index := int(time.Now().UnixNano()) % len(accounts)
	return accounts[index]
}

// validateAndFixThinkingParams validates and automatically fixes extended thinking parameters
// If max_tokens <= thinking.budget_tokens, it adjusts max_tokens to be budget_tokens + buffer
func (s *ProxyService) validateAndFixThinkingParams(bodyBytes []byte) ([]byte, error) {
	// Try to parse as JSON
	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		// Not JSON or invalid - let Claude API handle it
		return bodyBytes, nil
	}

	// Check if thinking mode is enabled
	thinking, hasThinking := body["thinking"].(map[string]interface{})
	if !hasThinking {
		// No thinking configuration - pass through
		return bodyBytes, nil
	}

	// Extract thinking.budget_tokens
	budgetTokensFloat, hasBudget := thinking["budget_tokens"].(float64)
	if !hasBudget {
		// No budget_tokens specified - pass through
		return bodyBytes, nil
	}
	budgetTokens := int(budgetTokensFloat)

	// Extract max_tokens
	maxTokensFloat, hasMaxTokens := body["max_tokens"].(float64)
	if !hasMaxTokens {
		// No max_tokens specified - pass through
		return bodyBytes, nil
	}
	maxTokens := int(maxTokensFloat)

	// Check if max_tokens > budget_tokens (Claude API requirement)
	if maxTokens > budgetTokens {
		// Valid configuration - pass through
		return bodyBytes, nil
	}

	// Invalid configuration detected - auto-fix by increasing max_tokens
	// Add reasonable buffer (10% of budget_tokens or minimum 1024 tokens)
	buffer := budgetTokens / 10
	if buffer < 1024 {
		buffer = 1024
	}
	newMaxTokens := budgetTokens + buffer

	// Log the auto-correction
	s.logger.Withs(sctx.Fields{
		"original_max_tokens":   maxTokens,
		"budget_tokens":         budgetTokens,
		"adjusted_max_tokens":   newMaxTokens,
		"buffer_added":          buffer,
	}).Warn("Auto-corrected max_tokens for extended thinking mode - max_tokens must be greater than budget_tokens")

	// Update max_tokens in the request body
	body["max_tokens"] = newMaxTokens

	// Re-serialize to JSON
	modifiedBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified request body: %w", err)
	}

	return modifiedBody, nil
}
