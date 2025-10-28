package repositories

import (
	"context"
	"fmt"
	"sync"

	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	sctx "github.com/phathdt/service-context"
)

// MemoryAccountRepository implements in-memory storage for accounts
type MemoryAccountRepository struct {
	accounts map[string]*entities.Account // accountID -> account
	mu       sync.RWMutex
	logger   sctx.Logger
}

// NewMemoryAccountRepository creates a new in-memory account repository
func NewMemoryAccountRepository(appLogger sctx.Logger) interfaces.AccountRepository {
	logger := appLogger.Withs(sctx.Fields{"component": "memory-account-repository"})

	return &MemoryAccountRepository{
		accounts: make(map[string]*entities.Account),
		logger:   logger,
	}
}

// Create stores a new account in memory
func (r *MemoryAccountRepository) Create(ctx context.Context, account *entities.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[account.ID]; exists {
		return fmt.Errorf("account already exists: %s", account.ID)
	}

	r.accounts[account.ID] = account
	r.logger.Withs(sctx.Fields{"account_id": account.ID}).Debug("Account created in memory")
	return nil
}

// GetByID retrieves an account by ID
func (r *MemoryAccountRepository) GetByID(ctx context.Context, id string) (*entities.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, exists := r.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found: %s", id)
	}

	return account, nil
}

// List retrieves all accounts
func (r *MemoryAccountRepository) List(ctx context.Context) ([]*entities.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]*entities.Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Update updates an existing account
func (r *MemoryAccountRepository) Update(ctx context.Context, account *entities.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[account.ID]; !exists {
		return fmt.Errorf("account not found: %s", account.ID)
	}

	r.accounts[account.ID] = account
	r.logger.Withs(sctx.Fields{"account_id": account.ID}).Debug("Account updated in memory")
	return nil
}

// Delete removes an account by ID
func (r *MemoryAccountRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[id]; !exists {
		return fmt.Errorf("account not found: %s", id)
	}

	delete(r.accounts, id)
	r.logger.Withs(sctx.Fields{"account_id": id}).Debug("Account deleted from memory")
	return nil
}

// GetActiveAccounts retrieves all active accounts
func (r *MemoryAccountRepository) GetActiveAccounts(ctx context.Context) ([]*entities.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]*entities.Account, 0)
	for _, account := range r.accounts {
		if account.IsActive() {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}
