package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// AccountRepository defines the interface for app account persistence
type AccountRepository interface {
	// Create creates a new app account
	Create(ctx context.Context, account *entities.Account) error

	// GetByID retrieves an app account by ID
	GetByID(ctx context.Context, id string) (*entities.Account, error)

	// List retrieves all app accounts
	List(ctx context.Context) ([]*entities.Account, error)

	// Update updates an existing app account
	Update(ctx context.Context, account *entities.Account) error

	// Delete deletes an app account by ID
	Delete(ctx context.Context, id string) error

	// GetActiveAccounts retrieves all active app accounts
	GetActiveAccounts(ctx context.Context) ([]*entities.Account, error)
}
