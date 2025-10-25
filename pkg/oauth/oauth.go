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
)

// Service handles OAuth 2.0 operations
type Service struct {
	clientID     string
	authorizeURL string
	tokenURL     string
	redirectURI  string
	httpClient   *http.Client
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
func NewService(clientID, authorizeURL, tokenURL, redirectURI string) *Service {
	return &Service{
		clientID:     clientID,
		authorizeURL: authorizeURL,
		tokenURL:     tokenURL,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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
func (s *Service) BuildAuthorizationURL(challenge *PKCEChallenge) string {
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", s.redirectURI)
	params.Set("state", challenge.State)
	params.Set("code_challenge", challenge.CodeChallenge)
	params.Set("code_challenge_method", "S256")

	return fmt.Sprintf("%s?%s", s.authorizeURL, params.Encode())
}

// ExchangeCodeForToken exchanges authorization code for access and refresh tokens
func (s *Service) ExchangeCodeForToken(ctx context.Context, code, codeVerifier string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", s.clientID)
	data.Set("redirect_uri", s.redirectURI)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", s.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.clientID)

	req, err := http.NewRequestWithContext(ctx, "POST", s.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

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
