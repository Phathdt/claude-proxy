package oauth

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

// Service handles OAuth 2.0 operations
type Service struct {
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

// NewService creates a new OAuth service
func NewService(clientID, authorizeURL, tokenURL, redirectURI, scope string) *Service {
	return &Service{
		clientID:     clientID,
		authorizeURL: authorizeURL,
		tokenURL:     tokenURL,
		redirectURI:  redirectURI,
		scope:        scope,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: sctx.GlobalLogger().GetLogger("oauth"),
	}
}

// GeneratePKCEChallenge generates PKCE code verifier and challenge
func (s *Service) GeneratePKCEChallenge() (*PKCEChallenge, error) {
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
func (s *Service) BuildAuthorizationURL(challenge *PKCEChallenge, organizationID string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", s.clientID)

	// Add organization_uuid as query parameter if provided
	if organizationID != "" {
		params.Set("organization_uuid", organizationID)
	}

	params.Set("redirect_uri", s.redirectURI)

	// Add scope if configured
	if s.scope != "" {
		params.Set("scope", s.scope)
	}

	params.Set("state", challenge.State)
	params.Set("code_challenge", challenge.CodeChallenge)
	params.Set("code_challenge_method", "S256")

	return fmt.Sprintf("%s?%s", s.authorizeURL, params.Encode())
}

// ExchangeCodeForToken exchanges authorization code for access and refresh tokens
// Code format: "auth_code#state" - the state is appended after # if present
func (s *Service) ExchangeCodeForToken(ctx context.Context, code, codeVerifier string) (*TokenResponse, error) {
	// Split code and state (Python format: code contains "auth_code#state")
	parts := strings.Split(code, "#")
	authCode := parts[0]
	var state string
	if len(parts) > 1 {
		state = parts[1]
	}

	// Build JSON payload
	payload := map[string]string{
		"code":          authCode,
		"grant_type":    "authorization_code",
		"client_id":     s.clientID,
		"redirect_uri":  s.redirectURI,
		"code_verifier": codeVerifier,
	}

	// Add state if present
	if state != "" {
		payload["state"] = state
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.tokenURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
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
func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	s.logger.Withs(sctx.Fields{
		"action": "refresh_token_start",
		"url":    s.tokenURL,
	}).Debug("Starting OAuth2 token refresh")

	// Build JSON payload (matching Python implementation)
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     s.clientID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"action": "refresh_token_error",
			"error":  err.Error(),
			"stage":  "marshal_request",
		}).Error("Failed to marshal refresh request")
		return nil, fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.tokenURL, strings.NewReader(string(jsonData)))
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"action": "refresh_token_error",
			"error":  err.Error(),
			"stage":  "create_request",
		}).Error("Failed to create refresh request")
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	s.logger.Withs(sctx.Fields{
		"action": "refresh_token_request_sent",
		"url":    s.tokenURL,
	}).Info("Sending token refresh request to OAuth server")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"action": "refresh_token_error",
			"error":  err.Error(),
			"stage":  "http_request",
		}).Error("Failed to refresh token (HTTP request failed)")
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"action":      "refresh_token_error",
			"error":       err.Error(),
			"stage":       "read_response",
			"status_code": resp.StatusCode,
		}).Error("Failed to read refresh response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the entire response for debugging
	s.logger.Withs(sctx.Fields{
		"action":      "refresh_token_response",
		"status_code": resp.StatusCode,
	}).Debug("=== OAUTH2 TOKEN REFRESH RESPONSE START ===")

	s.logger.Withs(sctx.Fields{
		"headers": resp.Header,
	}).Debug("Response Headers")

	s.logger.Withs(sctx.Fields{
		"body": string(body),
	}).Debug("Response Body")

	s.logger.Debug("=== OAUTH2 TOKEN REFRESH RESPONSE END ===")

	if resp.StatusCode != http.StatusOK {
		s.logger.Withs(sctx.Fields{
			"action":      "refresh_token_error",
			"status_code": resp.StatusCode,
			"error_body":  string(body),
			"stage":       "http_status",
		}).Error("Token refresh failed with non-200 status")
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		s.logger.Withs(sctx.Fields{
			"action":        "refresh_token_error",
			"error":         err.Error(),
			"stage":         "decode_response",
			"response_body": string(body),
		}).Error("Failed to decode refresh response")
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	s.logger.Withs(sctx.Fields{
		"action":      "refresh_token_success",
		"expires_in":  tokenResp.ExpiresIn,
		"token_type":  tokenResp.TokenType,
		"has_refresh": tokenResp.RefreshToken != "",
	}).Info("Successfully refreshed OAuth2 tokens")

	return &tokenResp, nil
}

// GetOrganizationUUID fetches the organization UUID from Claude API
func (s *Service) GetOrganizationUUID(ctx context.Context, accessToken, claudeBaseURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/organizations", claudeBaseURL), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create organizations request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get organizations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get organizations failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read organizations response: %w", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode organizations response: %w", err)
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("no organizations found for this account")
	}

	return result.Data[0].ID, nil
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}
