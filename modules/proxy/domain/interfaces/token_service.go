package interfaces

import (
	"context"

	"claude-proxy/modules/proxy/domain/entities"
)

// TokenService defines the interface for token management operations
type TokenService interface {
	// CreateToken creates a new API token
	CreateToken(ctx context.Context, name, key string, status entities.TokenStatus) (*entities.Token, error)

	// GetToken retrieves a token by ID
	GetToken(ctx context.Context, id string) (*entities.Token, error)

	// ListTokens retrieves all tokens
	ListTokens(ctx context.Context) ([]*entities.Token, error)

	// UpdateToken updates an existing token
	UpdateToken(ctx context.Context, id, name, key string, status entities.TokenStatus) (*entities.Token, error)

	// DeleteToken deletes a token
	DeleteToken(ctx context.Context, id string) error

	// ValidateToken validates a token and increments usage
	ValidateToken(ctx context.Context, key string) (*entities.Token, error)

	// GenerateKey generates a random API key
	GenerateKey() (string, error)
}
