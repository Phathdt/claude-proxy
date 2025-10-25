package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"claude-proxy/pkg/account"
	"claude-proxy/pkg/errors"
	"claude-proxy/pkg/token"

	"github.com/gin-gonic/gin"
	sctx "github.com/phathdt/service-context"
)

// ProxyHandler handles proxying requests to Claude API
type ProxyHandler struct {
	tokenManager    *token.Manager
	accountManager  *account.MultiAccountManager
	claudeBaseURL   string
	logger          sctx.Logger
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(
	tokenManager *token.Manager,
	accountManager *account.MultiAccountManager,
	claudeBaseURL string,
	logger sctx.Logger,
) *ProxyHandler {
	return &ProxyHandler{
		tokenManager:   tokenManager,
		accountManager: accountManager,
		claudeBaseURL:  claudeBaseURL,
		logger:         logger,
	}
}

// ProxyToClaudeAPI proxies requests to Claude API with token verification
func (h *ProxyHandler) ProxyToClaudeAPI(c *gin.Context) {
	// Extract bearer token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		panic(errors.NewUnauthorizedError("missing authorization header"))
	}

	// Parse bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		panic(errors.NewUnauthorizedError("invalid authorization header format"))
	}
	bearerToken := parts[1]

	// Validate token using token manager
	validatedToken, err := h.tokenManager.ValidateToken(bearerToken)
	if err != nil {
		h.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Warn("Token validation failed")
		panic(errors.NewUnauthorizedError("invalid or inactive token"))
	}

	h.logger.Withs(sctx.Fields{
		"token_id":   validatedToken.ID,
		"token_name": validatedToken.Name,
		"path":       c.Request.URL.Path,
	}).Info("Token validated successfully")

	// Get active app account
	accounts := h.accountManager.ListAccounts()
	var selectedAccount *account.AppAccount
	for _, acc := range accounts {
		if acc.Status == "active" {
			selectedAccount = acc
			break
		}
	}

	if selectedAccount == nil {
		h.logger.Error("No active app accounts found")
		panic(errors.NewServiceUnavailableError("no active claude accounts available"))
	}

	// Get valid access token (will auto-refresh if needed)
	ctx := context.Background()
	accessToken, err := h.accountManager.GetValidToken(ctx, selectedAccount.ID)
	if err != nil {
		h.logger.Withs(sctx.Fields{
			"error":      err.Error(),
			"account_id": selectedAccount.ID,
		}).Error("Failed to get valid access token")
		panic(errors.NewServiceUnavailableError("failed to get valid claude access token"))
	}

	h.logger.Withs(sctx.Fields{
		"account_id":   selectedAccount.ID,
		"account_name": selectedAccount.Name,
		"org_uuid":     selectedAccount.OrganizationUUID,
	}).Info("Using app account for proxy")

	// Build target URL
	targetURL := h.claudeBaseURL + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err != nil {
			h.logger.Withs(sctx.Fields{"error": err.Error()}).Error("Failed to read request body")
			panic(errors.NewBadRequestError("BAD_REQUEST", "Failed to read request body", err.Error()))
		}
	}

	// Create proxy request
	proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		h.logger.Withs(sctx.Fields{"error": err.Error()}).Error("Failed to create proxy request")
		panic(errors.NewInternalServerError("failed to create proxy request"))
	}

	// Copy headers from original request (except Authorization and Host)
	for key, values := range c.Request.Header {
		if key == "Authorization" || key == "Host" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Set Claude API authorization with app account's access token
	proxyReq.Header.Set("Authorization", "Bearer "+accessToken)

	// Set content type if not present
	if proxyReq.Header.Get("Content-Type") == "" && len(bodyBytes) > 0 {
		proxyReq.Header.Set("Content-Type", "application/json")
	}

	// Set anthropic version header if not present
	if proxyReq.Header.Get("anthropic-version") == "" {
		proxyReq.Header.Set("anthropic-version", "2023-06-01")
	}

	h.logger.Withs(sctx.Fields{
		"method":      c.Request.Method,
		"target_url":  targetURL,
		"token_id":    validatedToken.ID,
		"account_id":  selectedAccount.ID,
	}).Info("Proxying request to Claude API")

	// Send request to Claude API
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		h.logger.Withs(sctx.Fields{"error": err.Error()}).Error("Failed to proxy request")
		panic(errors.NewServiceUnavailableError("failed to proxy request to claude api"))
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Withs(sctx.Fields{"error": err.Error()}).Error("Failed to read response body")
		panic(errors.NewInternalServerError("failed to read claude api response"))
	}

	h.logger.Withs(sctx.Fields{
		"status_code": resp.StatusCode,
		"token_id":    validatedToken.ID,
		"account_id":  selectedAccount.ID,
	}).Info("Received response from Claude API")

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Return response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// GetModels handles GET /v1/models
func (h *ProxyHandler) GetModels(c *gin.Context) {
	h.ProxyToClaudeAPI(c)
}

// CreateMessage handles POST /v1/messages
func (h *ProxyHandler) CreateMessage(c *gin.Context) {
	h.ProxyToClaudeAPI(c)
}
