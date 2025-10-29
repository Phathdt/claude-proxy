package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// TokenPersistenceRepository defines the interface for durable token storage
// Implementation should prioritize data durability and persistence over speed
// All operations should persist to disk or permanent storage
type TokenPersistenceRepository interface {
	// SaveAll persists all tokens to durable storage (batch operation)
	SaveAll(ctx context.Context, tokens []*entities.Token) error

	// LoadAll loads all tokens from durable storage
	LoadAll(ctx context.Context) ([]*entities.Token, error)

	// Create creates and persists a new token
	Create(ctx context.Context, token *entities.Token) error

	// Update updates and persists an existing token
	Update(ctx context.Context, token *entities.Token) error

	// Delete deletes a token from persistent storage
	Delete(ctx context.Context, id string) error
}
