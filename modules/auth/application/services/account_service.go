package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"
	"claude-proxy/pkg/oauth"

	"github.com/google/uuid"
	sctx "github.com/phathdt/service-context"
)

// AccountService implements account management with hybrid storage
type AccountService struct {
	memoryRepo   interfaces.AccountRepository
	jsonRepo     interfaces.AccountRepository
	oauthService *oauth.Service
	dirty        bool
	mu           sync.RWMutex
	logger       sctx.Logger
}

// NewAccountService creates a new account service
func NewAccountService(
	memoryRepo interfaces.AccountRepository,
	jsonRepo interfaces.AccountRepository,
	oauthService *oauth.Service,
	appLogger sctx.Logger,
) interfaces.AccountService {
	logger := appLogger.Withs(sctx.Fields{"component": "account-service"})

	svc := &AccountService{
		memoryRepo:   memoryRepo,
		jsonRepo:     jsonRepo,
		oauthService: oauthService,
		dirty:        false,
		logger:       logger,
	}

	// Load from JSON into memory on init
	if err := svc.loadFromJSON(); err != nil {
		logger.Withs(sctx.Fields{"error": err}).Warn("Failed to load accounts from JSON")
	}

	return svc
}

// loadFromJSON loads all accounts from JSON into memory
func (s *AccountService) loadFromJSON() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accounts, err := s.jsonRepo.List(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list accounts from JSON: %w", err)
	}

	// Load each account into memory
	for _, account := range accounts {
		if err := s.memoryRepo.Create(context.Background(), account); err != nil {
			s.logger.Withs(sctx.Fields{
				"account_id": account.ID,
				"error":      err,
			}).Warn("Failed to load account into memory")
		}
	}

	s.logger.Withs(sctx.Fields{"count": len(accounts)}).Info("Accounts loaded from JSON to memory")
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

// Sync syncs in-memory data to JSON (called every 1 minute)
func (s *AccountService) Sync(ctx context.Context) error {
	if !s.isDirty() {
		return nil // No changes, skip sync
	}

	s.logger.Debug("Syncing accounts to JSON")

	// Get all accounts from memory
	accounts, err := s.memoryRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list accounts from memory: %w", err)
	}

	// Sync each account to JSON
	for _, account := range accounts {
		// Check if exists in JSON
		existing, err := s.jsonRepo.GetByID(ctx, account.ID)
		if err != nil {
			// Doesn't exist, create
			if err := s.jsonRepo.Create(ctx, account); err != nil {
				s.logger.Withs(sctx.Fields{
					"account_id": account.ID,
					"error":      err,
				}).Error("Failed to create account in JSON")
				continue
			}
		} else if existing != nil {
			// Exists, update
			if err := s.jsonRepo.Update(ctx, account); err != nil {
				s.logger.Withs(sctx.Fields{
					"account_id": account.ID,
					"error":      err,
				}).Error("Failed to update account in JSON")
				continue
			}
		}
	}

	s.clearDirty()
	s.logger.Withs(sctx.Fields{"count": len(accounts)}).Info("Accounts synced to JSON")
	return nil
}

// FinalSync performs final sync on graceful shutdown
func (s *AccountService) FinalSync(ctx context.Context) error {
	s.logger.Info("Performing final sync of accounts")
	return s.Sync(ctx)
}

// CreateAccount creates a new account from OAuth code
func (s *AccountService) CreateAccount(ctx context.Context, name, code, codeVerifier, orgID string) (*entities.Account, error) {
	// Exchange code for tokens using PKCE code verifier
	tokenResp, err := s.oauthService.ExchangeCodeForToken(ctx, code, codeVerifier)
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

	// Save to memory
	if err := s.memoryRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": account.ID, "name": name}).Info("Account created")

	return account, nil
}

// GetAccount retrieves account by ID
func (s *AccountService) GetAccount(ctx context.Context, id string) (*entities.Account, error) {
	return s.memoryRepo.GetByID(ctx, id)
}

// ListAccounts retrieves all accounts
func (s *AccountService) ListAccounts(ctx context.Context) ([]*entities.Account, error) {
	return s.memoryRepo.List(ctx)
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(
	ctx context.Context,
	id, name string,
	status entities.AccountStatus,
) (*entities.Account, error) {
	account, err := s.memoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	account.Update(name, status)

	if err := s.memoryRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": id}).Info("Account updated")
	return account, nil
}

// DeleteAccount deletes an account
func (s *AccountService) DeleteAccount(ctx context.Context, id string) error {
	if err := s.memoryRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": id}).Info("Account deleted")
	return nil
}

// GetActiveAccounts retrieves all active accounts
func (s *AccountService) GetActiveAccounts(ctx context.Context) ([]*entities.Account, error) {
	return s.memoryRepo.GetActiveAccounts(ctx)
}

// GetValidToken returns a valid access token for an account (with auto-refresh)
func (s *AccountService) GetValidToken(ctx context.Context, accountID string) (string, error) {
	account, err := s.memoryRepo.GetByID(ctx, accountID)
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
	tokenResp, err := s.oauthService.RefreshAccessToken(ctx, account.RefreshToken)
	if err != nil {
		account.UpdateRefreshError(err.Error())
		s.memoryRepo.Update(ctx, account)
		s.markDirty()
		return err
	}

	account.UpdateTokens(tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn)
	if err := s.memoryRepo.Update(ctx, account); err != nil {
		return err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"account_id": account.ID}).Info("Token refreshed")
	return nil
}

// RefreshAllAccounts refreshes tokens for all accounts that need it
func (s *AccountService) RefreshAllAccounts(ctx context.Context) (int, int, int, error) {
	accounts, err := s.memoryRepo.GetActiveAccounts(ctx)
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
	accounts, err := s.memoryRepo.List(ctx)
	if err != nil {
		return 0, err
	}

	recovered := 0
	for _, account := range accounts {
		if account.Status == entities.AccountStatusRateLimited && account.IsRateLimitExpired() {
			account.RecoverFromRateLimit()
			if err := s.memoryRepo.Update(ctx, account); err != nil {
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
	accounts, err := s.memoryRepo.List(ctx)
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
