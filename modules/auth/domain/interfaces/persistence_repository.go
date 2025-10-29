package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// PersistenceRepository defines the interface for durable account storage
// Implementation should prioritize data durability and persistence over speed
// All operations should persist to disk or permanent storage
type PersistenceRepository interface {
	// SaveAll persists all accounts to durable storage (batch operation)
	SaveAll(ctx context.Context, accounts []*entities.Account) error

	// LoadAll loads all accounts from durable storage
	LoadAll(ctx context.Context) ([]*entities.Account, error)

	// Create creates and persists a new account
	Create(ctx context.Context, account *entities.Account) error

	// Update updates and persists an existing account
	Update(ctx context.Context, account *entities.Account) error

	// Delete deletes an account from persistent storage
	Delete(ctx context.Context, id string) error
}
