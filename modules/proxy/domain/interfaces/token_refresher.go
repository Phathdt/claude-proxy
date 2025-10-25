package interfaces

import "context"

// TokenRefresher defines the interface for refreshing OAuth tokens
type TokenRefresher interface {
	// RefreshAccessToken refreshes an OAuth access token
	// Returns: new access token, new refresh token, expires in seconds, error
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, int, error)
}
