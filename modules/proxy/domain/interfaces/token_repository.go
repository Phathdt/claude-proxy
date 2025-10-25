package interfaces

import (
	"context"

	"claude-proxy/modules/proxy/domain/entities"
)

// TokenRepository defines the interface for token persistence
type TokenRepository interface {
	// Create creates a new token
	Create(ctx context.Context, token *entities.Token) error

	// GetByID retrieves a token by ID
	GetByID(ctx context.Context, id string) (*entities.Token, error)

	// GetByKey retrieves a token by its key
	GetByKey(ctx context.Context, key string) (*entities.Token, error)

	// List retrieves all tokens
	List(ctx context.Context) ([]*entities.Token, error)

	// Update updates an existing token
	Update(ctx context.Context, token *entities.Token) error

	// Delete deletes a token by ID
	Delete(ctx context.Context, id string) error
}
