package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// AccountService defines the interface for app account management operations
type AccountService interface {
	// CreateAccount creates a new app account from OAuth code
	CreateAccount(
		ctx context.Context,
		name, code string,
	) (*entities.Account, error)

	// GetAccount retrieves an account by ID
	GetAccount(ctx context.Context, id string) (*entities.Account, error)

	// ListAccounts retrieves all accounts
	ListAccounts(ctx context.Context) ([]*entities.Account, error)

	// UpdateAccount updates an existing account
	UpdateAccount(ctx context.Context, id, name string, status entities.AccountStatus) (*entities.Account, error)

	// DeleteAccount deletes an account
	DeleteAccount(ctx context.Context, id string) error

	// GetValidToken returns a valid access token for an account, refreshing if needed
	GetValidToken(ctx context.Context, accountID string) (string, error)

	// GetActiveAccounts retrieves all active accounts
	GetActiveAccounts(ctx context.Context) ([]*entities.Account, error)

	// RefreshAllAccounts refreshes tokens for all accounts that need it
	// Returns refreshed count, failed count, skipped count, and error
	RefreshAllAccounts(ctx context.Context) (int, int, int, error)

	// RecoverRateLimitedAccounts checks and recovers accounts with expired rate limits
	// Returns the number of accounts recovered
	RecoverRateLimitedAccounts(ctx context.Context) (int, error)

	// GetStatistics returns system statistics including account counts and health metrics
	GetStatistics(ctx context.Context) (map[string]interface{}, error)

	// Sync syncs in-memory data to persistent storage
	Sync(ctx context.Context) error

	// FinalSync performs final sync on graceful shutdown
	FinalSync(ctx context.Context) error
}
