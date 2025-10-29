package interfaces

import (
	"context"

	"claude-proxy/modules/auth/infrastructure/clients"
)

// OAuthClient defines the interface for OAuth 2.0 operations
type OAuthClient interface {
	// GeneratePKCEChallenge generates PKCE code verifier and challenge
	GeneratePKCEChallenge() (*clients.PKCEChallenge, error)

	// BuildAuthorizationURL builds the OAuth authorization URL
	BuildAuthorizationURL(challenge *clients.PKCEChallenge, organizationID string) string

	// ExchangeCodeForToken exchanges authorization code for access and refresh tokens
	ExchangeCodeForToken(ctx context.Context, code, codeVerifier string) (*clients.TokenResponse, error)

	// RefreshAccessToken uses refresh token to get a new access token
	RefreshAccessToken(ctx context.Context, refreshToken string) (*clients.TokenResponse, error)
}
