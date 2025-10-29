package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// CacheRepository defines the interface for fast, volatile account storage
// Implementation should prioritize speed over durability
type CacheRepository interface {
	// Create creates a new account in cache
	Create(ctx context.Context, account *entities.Account) error

	// GetByID retrieves an account by ID from cache
	GetByID(ctx context.Context, id string) (*entities.Account, error)

	// List retrieves all accounts from cache
	List(ctx context.Context) ([]*entities.Account, error)

	// Update updates an existing account in cache
	Update(ctx context.Context, account *entities.Account) error

	// Delete deletes an account by ID from cache
	Delete(ctx context.Context, id string) error

	// GetActiveAccounts retrieves all active accounts from cache
	GetActiveAccounts(ctx context.Context) ([]*entities.Account, error)
}
