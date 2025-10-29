package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	"github.com/google/uuid"
	sctx "github.com/phathdt/service-context"
)

// AccountService implements account management with hybrid storage pattern
// Uses CacheRepository for fast in-memory access and PersistenceRepository for durability
type AccountService struct {
	cacheRepo       interfaces.CacheRepository
	persistenceRepo interfaces.PersistenceRepository
	oauthClient     interfaces.OAuthClient
	dirty           bool
	mu              sync.RWMutex
	logger          sctx.Logger
}

// NewAccountService creates a new account service with cache and persistence layers
func NewAccountService(
	cacheRepo interfaces.CacheRepository,
	persistenceRepo interfaces.PersistenceRepository,
	oauthClient interfaces.OAuthClient,
	appLogger sctx.Logger,
) interfaces.AccountService {
	logger := appLogger.Withs(sctx.Fields{"component": "account-service"})

	svc := &AccountService{
		cacheRepo:       cacheRepo,
		persistenceRepo: persistenceRepo,
		oauthClient:     oauthClient,
		dirty:           false,
		logger:          logger,
	}

	// Load from persistent storage into cache on init
	if err := svc.loadFromPersistence(); err != nil {
		logger.Withs(sctx.Fields{"error": err}).Warn("Failed to load accounts from persistence")
	}

	return svc
}

// loadFromPersistence loads all accounts from persistent storage into cache
func (s *AccountService) loadFromPersistence() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accounts, err := s.persistenceRepo.LoadAll(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load accounts from persistence: %w", err)
	}

	// Load each account into cache
	for _, account := range accounts {
		if err := s.cacheRepo.Create(context.Background(), account); err != nil {
			s.logger.Withs(sctx.Fields{
				"account_id": account.ID,
				"error":      err,
			}).Warn("Failed to load account into cache")
		}
	}

	s.logger.Withs(sctx.Fields{"count": len(accounts)}).Info("Accounts loaded from persistence to cache")
	return nil
}

// markDirty marks data as changed
func (s *AccountService) markDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = true
}

// isDirty checks if data has changed
func (s *AccountService) isDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

// clearDirty clears the dirty flag
func (s *AccountService) clearDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = false
}

// Sync syncs cache data to persistent storage (called every 1 minute)
func (s *AccountService) Sync(ctx context.Context) error {
	if !s.isDirty() {
		return nil // No changes, skip sync
	}

	s.logger.Debug("Syncing accounts to persistent storage")

	// Get all accounts from cache
	accounts, err := s.cacheRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list accounts from cache: %w", err)
	}

	// Batch save all accounts to persistent storage
	if err := s.persistenceRepo.SaveAll(ctx, accounts); err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err,
		}).Error("Failed to save accounts to persistence")
		return fmt.Errorf("failed to save accounts: %w", err)
	}

	s.clearDirty()
	s.logger.Withs(sctx.Fields{"count": len(accounts)}).Info("Accounts synced to persistent storage")
	return nil
}

// FinalSync performs final sync on graceful shutdown
func (s *AccountService) FinalSync(ctx context.Context) error {
	s.logger.Info("Performing final sync of accounts")
	return s.Sync(ctx)
}

// CreateAccount creates a new account from OAuth code
func (s *AccountService) CreateAccount(
	ctx context.Context,
	name, code, codeVerifier, orgID string,
) (*entities.Account, error) {
	// Exchange code for tokens using PKCE code verifier
	tokenResp, err := s.oauthClient.ExchangeCodeForToken(ctx, code, codeVerifier)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Use the provided organization UUID from the request
	// The org_id is included in the OAuth authorization URL and passed through the flow
	orgUUID := orgID

	// Create account entity
	now := time.Now()
	account := &entities.Account{
		ID:               uuid.Must(uuid.NewV7()).String(),
		Name:             name,
		OrganizationUUID: orgUUID,
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		ExpiresAt:        now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		RefreshAt:        now,
		Status:           entities.AccountStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Save to cache
	if err := s.cacheRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": account.ID, "name": name}).Info("Account created")

	return account, nil
}

// GetAccount retrieves account by ID
func (s *AccountService) GetAccount(ctx context.Context, id string) (*entities.Account, error) {
	return s.cacheRepo.GetByID(ctx, id)
}

// ListAccounts retrieves all accounts
func (s *AccountService) ListAccounts(ctx context.Context) ([]*entities.Account, error) {
	return s.cacheRepo.List(ctx)
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(
	ctx context.Context,
	id, name string,
	status entities.AccountStatus,
) (*entities.Account, error) {
	account, err := s.cacheRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	account.Update(name, status)

	if err := s.cacheRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": id}).Info("Account updated")
	return account, nil
}

// DeleteAccount deletes an account
func (s *AccountService) DeleteAccount(ctx context.Context, id string) error {
	if err := s.cacheRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": id}).Info("Account deleted")
	return nil
}

// GetActiveAccounts retrieves all active accounts
func (s *AccountService) GetActiveAccounts(ctx context.Context) ([]*entities.Account, error) {
	return s.cacheRepo.GetActiveAccounts(ctx)
}

// GetValidToken returns a valid access token for an account (with auto-refresh)
func (s *AccountService) GetValidToken(ctx context.Context, accountID string) (string, error) {
	account, err := s.cacheRepo.GetByID(ctx, accountID)
	if err != nil {
		return "", err
	}

	// Check if needs refresh
	if account.NeedsRefresh() {
		if err := s.refreshToken(ctx, account); err != nil {
			s.logger.Withs(sctx.Fields{
				"account_id": accountID,
				"error":      err,
			}).Warn("Failed to refresh token")
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	return account.AccessToken, nil
}

// refreshToken refreshes account tokens
func (s *AccountService) refreshToken(ctx context.Context, account *entities.Account) error {
	tokenResp, err := s.oauthClient.RefreshAccessToken(ctx, account.RefreshToken)
	if err != nil {
		account.UpdateRefreshError(err.Error())
		s.cacheRepo.Update(ctx, account)
		s.markDirty()
		return err
	}

	account.UpdateTokens(tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn)
	if err := s.cacheRepo.Update(ctx, account); err != nil {
		return err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": account.ID}).Info("Token refreshed")
	return nil
}

// RefreshAllAccounts refreshes tokens for all accounts that need it
func (s *AccountService) RefreshAllAccounts(ctx context.Context) (int, int, int, error) {
	accounts, err := s.cacheRepo.GetActiveAccounts(ctx)
	if err != nil {
		return 0, 0, 0, err
	}

	refreshed, failed, skipped := 0, 0, 0

	for _, account := range accounts {
		if !account.NeedsRefresh() {
			skipped++
			continue
		}

		if err := s.refreshToken(ctx, account); err != nil {
			failed++
			continue
		}

		refreshed++
	}

	return refreshed, failed, skipped, nil
}

// RecoverRateLimitedAccounts checks and recovers accounts with expired rate limits
func (s *AccountService) RecoverRateLimitedAccounts(ctx context.Context) (int, error) {
	accounts, err := s.cacheRepo.List(ctx)
	if err != nil {
		return 0, err
	}

	recovered := 0
	for _, account := range accounts {
		if account.Status == entities.AccountStatusRateLimited && account.IsRateLimitExpired() {
			account.RecoverFromRateLimit()
			if err := s.cacheRepo.Update(ctx, account); err != nil {
				s.logger.Withs(sctx.Fields{
					"account_id": account.ID,
					"error":      err,
				}).Warn("Failed to recover rate limited account")
				continue
			}
			s.markDirty()
			recovered++
		}
	}

	return recovered, nil
}

// GetStatistics returns system statistics including account counts and health metrics
func (s *AccountService) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	accounts, err := s.cacheRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Count by status
	activeCount := 0
	inactiveCount := 0
	rateLimitedCount := 0
	invalidCount := 0
	needsRefreshCount := 0

	var oldestTokenAge time.Duration
	now := time.Now()

	for _, account := range accounts {
		switch account.Status {
		case entities.AccountStatusActive:
			activeCount++
		case entities.AccountStatusInactive:
			inactiveCount++
		case entities.AccountStatusRateLimited:
			rateLimitedCount++
		case entities.AccountStatusInvalid:
			invalidCount++
		}

		// Check if account needs refresh (within 60s of expiry)
		if account.NeedsRefresh() {
			needsRefreshCount++
		}

		// Track oldest token age
		tokenAge := now.Sub(account.ExpiresAt.Add(-1 * time.Hour)) // Tokens valid for 1 hour
		if tokenAge > oldestTokenAge {
			oldestTokenAge = tokenAge
		}
	}

	// Calculate system health
	systemHealth := "healthy"
	if invalidCount > 0 || rateLimitedCount > len(accounts)/2 {
		systemHealth = "unhealthy"
	} else if rateLimitedCount > 0 || needsRefreshCount > len(accounts)/2 {
		systemHealth = "degraded"
	}

	stats := make(map[string]interface{})
	stats["total_accounts"] = len(accounts)
	stats["active_accounts"] = activeCount
	stats["inactive_accounts"] = inactiveCount
	stats["rate_limited_accounts"] = rateLimitedCount
	stats["invalid_accounts"] = invalidCount
	stats["accounts_needing_refresh"] = needsRefreshCount
	stats["oldest_token_age_hours"] = oldestTokenAge.Hours()
	stats["system_health"] = systemHealth

	return stats, nil
}
