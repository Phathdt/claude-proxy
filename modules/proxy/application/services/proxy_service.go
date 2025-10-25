package services

import (
	"context"
	"fmt"
	"io"
	"net/http"

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
func (s *ProxyService) ProxyRequest(ctx context.Context, token *entities.Token, req *http.Request) (*http.Response, error) {
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

	// Build path with query string
	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}

	// Prepare headers
	headers := make(map[string]string)
	for key, values := range req.Header {
		if key == "Authorization" || key == "Host" {
			continue // Skip these headers
		}
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Set Claude API authorization
	headers["Authorization"] = "Bearer " + accessToken

	// Proxy the request
	resp, err := s.claudeClient.ProxyRequest(ctx, req.Method, path, headers, bodyBytes)
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

// GetValidAccount returns a valid active account with a fresh access token
func (s *ProxyService) GetValidAccount(ctx context.Context) (*entities.Account, error) {
	accounts, err := s.accountRepo.GetActiveAccounts(ctx)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no active accounts available")
	}

	// Return first active account
	// TODO: Implement load balancing or selection strategy
	return accounts[0], nil
}
