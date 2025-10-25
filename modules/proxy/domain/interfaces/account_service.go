package interfaces

import (
	"context"

	"claude-proxy/modules/proxy/domain/entities"
)

// AccountService defines the interface for app account management operations
type AccountService interface {
	// CreateAccount creates a new app account
	CreateAccount(ctx context.Context, name, orgUUID, accessToken, refreshToken string, expiresIn int) (*entities.AppAccount, error)

	// GetAccount retrieves an account by ID
	GetAccount(ctx context.Context, id string) (*entities.AppAccount, error)

	// ListAccounts retrieves all accounts
	ListAccounts(ctx context.Context) ([]*entities.AppAccount, error)

	// UpdateAccount updates an existing account
	UpdateAccount(ctx context.Context, id, name string, status entities.AccountStatus) (*entities.AppAccount, error)

	// DeleteAccount deletes an account
	DeleteAccount(ctx context.Context, id string) error

	// GetValidToken returns a valid access token for an account, refreshing if needed
	GetValidToken(ctx context.Context, accountID string) (string, error)
}
