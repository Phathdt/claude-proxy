package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// TokenCacheRepository defines the interface for fast, volatile token storage
// Implementation should prioritize speed over durability
type TokenCacheRepository interface {
	// Create creates a new token in cache
	Create(ctx context.Context, token *entities.Token) error

	// GetByID retrieves a token by ID from cache
	GetByID(ctx context.Context, id string) (*entities.Token, error)

	// GetByKey retrieves a token by its key from cache
	GetByKey(ctx context.Context, key string) (*entities.Token, error)

	// List retrieves all tokens from cache
	List(ctx context.Context) ([]*entities.Token, error)

	// Update updates an existing token in cache
	Update(ctx context.Context, token *entities.Token) error

	// Delete deletes a token by ID from cache
	Delete(ctx context.Context, id string) error
}
