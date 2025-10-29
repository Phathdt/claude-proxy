package clients

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	sctx "github.com/phathdt/service-context"
)

// OAuthClient handles OAuth 2.0 operations for Claude authentication
type OAuthClient struct {
	clientID     string
	authorizeURL string
	tokenURL     string
	redirectURI  string
	scope        string
	httpClient   *http.Client
	logger       sctx.Logger
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// PKCEChallenge holds PKCE challenge data
type PKCEChallenge struct {
	CodeVerifier  string
	CodeChallenge string
	State         string
}

// NewOAuthClient creates a new OAuth client for Claude authentication
func NewOAuthClient(clientID, authorizeURL, tokenURL, redirectURI, scope string, logger sctx.Logger) *OAuthClient {
	return &OAuthClient{
		clientID:     clientID,
		authorizeURL: authorizeURL,
		tokenURL:     tokenURL,
		redirectURI:  redirectURI,
		scope:        scope,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GeneratePKCEChallenge generates PKCE code verifier and challenge
func (c *OAuthClient) GeneratePKCEChallenge() (*PKCEChallenge, error) {
	// Generate code verifier (43-128 characters)
	verifier, err := generateRandomString(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Generate code challenge (SHA256 hash of verifier)
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// Generate state for CSRF protection
	state, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	return &PKCEChallenge{
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
		State:         state,
	}, nil
}

// BuildAuthorizationURL builds the OAuth authorization URL
// If organizationID is provided, it's added as organization_uuid query parameter
func (c *OAuthClient) BuildAuthorizationURL(challenge *PKCEChallenge, organizationID string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", c.clientID)

	// Add organization_uuid as query parameter if provided
	if organizationID != "" {
		params.Set("organization_uuid", organizationID)
	}

	params.Set("redirect_uri", c.redirectURI)

	// Add scope if configured
	if c.scope != "" {
		params.Set("scope", c.scope)
	}

	params.Set("state", challenge.State)
	params.Set("code_challenge", challenge.CodeChallenge)
	params.Set("code_challenge_method", "S256")

	return fmt.Sprintf("%s?%s", c.authorizeURL, params.Encode())
}

// ExchangeCodeForToken exchanges authorization code for access and refresh tokens
// Code format: "auth_code#state" - the state is appended after # if present
func (c *OAuthClient) ExchangeCodeForToken(ctx context.Context, code, codeVerifier string) (*TokenResponse, error) {
	// Split code and state (Python format: code contains "auth_code#state")
	parts := strings.Split(code, "#")
	authCode := parts[0]
	var state string
	if len(parts) > 1 {
		state = parts[1]
	}

	// Build JSON payload (Claude's OAuth API uses JSON, not form-urlencoded)
	// Note: state parameter IS required for Claude's OAuth implementation
	payload := map[string]string{
		"code":          authCode,
		"grant_type":    "authorization_code",
		"client_id":     c.clientID,
		"redirect_uri":  c.redirectURI,
		"code_verifier": codeVerifier,
	}

	// Add state if present (Claude requires this)
	if state != "" {
		payload["state"] = state
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	// Debug logging
	c.logger.Withs(sctx.Fields{
		"url":     c.tokenURL,
		"payload": string(jsonData),
	}).Info("Sending token exchange request to Claude OAuth API")

	req, err := http.NewRequestWithContext(ctx, "POST", c.tokenURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshAccessToken uses refresh token to get a new access token
func (c *OAuthClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	c.logger.Withs(sctx.Fields{
		"action": "refresh_token_start",
		"url":    c.tokenURL,
	}).Debug("Starting OAuth2 token refresh")

	// Build JSON payload (matching Python implementation)
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     c.clientID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.logger.Withs(sctx.Fields{
			"action": "refresh_token_error",
			"error":  err.Error(),
			"stage":  "marshal_request",
		}).Error("Failed to marshal refresh request")
		return nil, fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.tokenURL, strings.NewReader(string(jsonData)))
	if err != nil {
		c.logger.Withs(sctx.Fields{
			"action": "refresh_token_error",
			"error":  err.Error(),
			"stage":  "create_request",
		}).Error("Failed to create refresh request")
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Withs(sctx.Fields{
		"action": "refresh_token_request_sent",
		"url":    c.tokenURL,
	}).Info("Sending token refresh request to OAuth server")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Withs(sctx.Fields{
			"action": "refresh_token_error",
			"error":  err.Error(),
			"stage":  "http_request",
		}).Error("Failed to refresh token (HTTP request failed)")
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Withs(sctx.Fields{
			"action":      "refresh_token_error",
			"error":       err.Error(),
			"stage":       "read_response",
			"status_code": resp.StatusCode,
		}).Error("Failed to read refresh response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the entire response for debugging
	c.logger.Withs(sctx.Fields{
		"action":      "refresh_token_response",
		"status_code": resp.StatusCode,
	}).Debug("=== OAUTH2 TOKEN REFRESH RESPONSE START ===")

	c.logger.Withs(sctx.Fields{
		"headers": resp.Header,
	}).Debug("Response Headers")

	c.logger.Withs(sctx.Fields{
		"body": string(body),
	}).Debug("Response Body")

	c.logger.Debug("=== OAUTH2 TOKEN REFRESH RESPONSE END ===")

	if resp.StatusCode != http.StatusOK {
		c.logger.Withs(sctx.Fields{
			"action":      "refresh_token_error",
			"status_code": resp.StatusCode,
			"error_body":  string(body),
			"stage":       "http_status",
		}).Error("Token refresh failed with non-200 status")
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		c.logger.Withs(sctx.Fields{
			"action":        "refresh_token_error",
			"error":         err.Error(),
			"stage":         "decode_response",
			"response_body": string(body),
		}).Error("Failed to decode refresh response")
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	c.logger.Withs(sctx.Fields{
		"action":      "refresh_token_success",
		"expires_in":  tokenResp.ExpiresIn,
		"token_type":  tokenResp.TokenType,
		"has_refresh": tokenResp.RefreshToken != "",
	}).Info("Successfully refreshed OAuth2 tokens")

	return &tokenResp, nil
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}
