package interfaces

import (
	"context"
	"net/http"

	"claude-proxy/modules/proxy/domain/entities"
)

// ProxyService defines the interface for proxy operations
type ProxyService interface {
	// ProxyRequest proxies an HTTP request to Claude API
	// It validates the token, selects an active account, and forwards the request
	ProxyRequest(ctx context.Context, token *entities.Token, req *http.Request) (*http.Response, error)

	// GetValidAccount returns a valid active account with a fresh access token
	GetValidAccount(ctx context.Context) (*entities.AppAccount, error)
}
