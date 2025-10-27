package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"

	sctx "github.com/phathdt/service-context"
)

// AccountService implements the account management business logic
type AccountService struct {
	accountRepo interfaces.AccountRepository
	refresher   interfaces.TokenRefresher
	logger      sctx.Logger
}

// NewAccountService creates a new account service
func NewAccountService(
	accountRepo interfaces.AccountRepository,
	refresher interfaces.TokenRefresher,
	logger sctx.Logger,
) interfaces.AccountService {
	return &AccountService{
		accountRepo: accountRepo,
		refresher:   refresher,
		logger:      logger,
	}
}

// CreateAccount creates a new app account
func (s *AccountService) CreateAccount(
	ctx context.Context,
	name, orgUUID, accessToken, refreshToken string,
	expiresIn int,
) (*entities.Account, error) {
	if name == "" {
		return nil, fmt.Errorf("account name is required")
	}

	now := time.Now()
	account := &entities.Account{
		Name:             name,
		OrganizationUUID: orgUUID,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        now.Add(time.Duration(expiresIn) * time.Second),
		Status:           entities.AccountStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.logger.Withs(sctx.Fields{
		"account_id": account.ID,
		"name":       account.Name,
		"org_uuid":   account.OrganizationUUID,
	}).Info("Account created successfully")

	return account, nil
}

// GetAccount retrieves an account by ID
func (s *AccountService) GetAccount(ctx context.Context, id string) (*entities.Account, error) {
	return s.accountRepo.GetByID(ctx, id)
}

// ListAccounts retrieves all accounts
func (s *AccountService) ListAccounts(ctx context.Context) ([]*entities.Account, error) {
	return s.accountRepo.List(ctx)
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(
	ctx context.Context,
	id, name string,
	status entities.AccountStatus,
) (*entities.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	account.Update(name, status)

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	s.logger.Withs(sctx.Fields{
		"account_id": account.ID,
		"name":       account.Name,
		"status":     account.Status,
	}).Info("Account updated successfully")

	return account, nil
}

// DeleteAccount deletes an account
func (s *AccountService) DeleteAccount(ctx context.Context, id string) error {
	if err := s.accountRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Withs(sctx.Fields{
		"account_id": id,
	}).Info("Account deleted successfully")

	return nil
}

// GetValidToken returns a valid access token for an account, refreshing if needed
func (s *AccountService) GetValidToken(ctx context.Context, accountID string) (string, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return "", err
	}

	// Check if token needs refresh (60s buffer)
	if account.NeedsRefresh() {
		s.logger.Withs(sctx.Fields{
			"account_id": accountID,
		}).Info("Token needs refresh, refreshing...")

		if err := s.refreshToken(ctx, account); err != nil {
			account.Deactivate()
			_ = s.accountRepo.Update(ctx, account)
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	return account.AccessToken, nil
}

// refreshToken refreshes an account's access token
func (s *AccountService) refreshToken(ctx context.Context, account *entities.Account) error {
	accessToken, refreshToken, expiresIn, err := s.refresher.RefreshAccessToken(ctx, account.RefreshToken)
	if err != nil {
		// Detect error type and update account status accordingly
		s.handleRefreshError(ctx, account, err)
		return err
	}

	// Success - clear any error state and mark as active
	account.UpdateTokens(accessToken, refreshToken, expiresIn)

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return err
	}

	s.logger.Withs(sctx.Fields{
		"account_id": account.ID,
		"expires_at": account.ExpiresAt,
	}).Info("Token refreshed successfully")

	return nil
}

// handleRefreshError analyzes refresh error and updates account status
func (s *AccountService) handleRefreshError(ctx context.Context, account *entities.Account, err error) {
	errMsg := err.Error()

	// Check for rate limit error (429 status code)
	if strings.Contains(errMsg, "429") || strings.Contains(strings.ToLower(errMsg), "rate limit") {
		// Rate limited - set 1 hour default recovery time
		until := time.Now().Add(1 * time.Hour)

		account.MarkRateLimited(until, errMsg)

		s.logger.Withs(sctx.Fields{
			"account_id":         account.ID,
			"rate_limited_until": until,
		}).Warn("Account marked as rate limited")

		_ = s.accountRepo.Update(ctx, account)
		return
	}

	// Check for invalid token error (401, 403 status codes)
	if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") ||
		strings.Contains(strings.ToLower(errMsg), "unauthorized") ||
		strings.Contains(strings.ToLower(errMsg), "invalid") {

		account.MarkInvalid(errMsg)

		s.logger.Withs(sctx.Fields{
			"account_id": account.ID,
			"error":      errMsg,
		}).Error("Account marked as invalid - credentials revoked or expired")

		_ = s.accountRepo.Update(ctx, account)
		return
	}

	// Generic error - mark as inactive but preserve error message
	account.UpdateRefreshError(errMsg)
	account.Deactivate()

	s.logger.Withs(sctx.Fields{
		"account_id": account.ID,
		"error":      errMsg,
	}).Error("Account refresh failed with unknown error")

	_ = s.accountRepo.Update(ctx, account)
}

// RecoverRateLimitedAccounts checks and recovers accounts with expired rate limits
// This should be called periodically by the scheduler
func (s *AccountService) RecoverRateLimitedAccounts(ctx context.Context) (int, error) {
	accounts, err := s.accountRepo.List(ctx)
	if err != nil {
		return 0, err
	}

	recoveredCount := 0

	for _, account := range accounts {
		if account.Status == entities.AccountStatusRateLimited && account.IsRateLimitExpired() {
			account.RecoverFromRateLimit()

			if err := s.accountRepo.Update(ctx, account); err != nil {
				s.logger.Withs(sctx.Fields{
					"account_id": account.ID,
					"error":      err.Error(),
				}).Error("Failed to recover rate limited account")
				continue
			}

			s.logger.Withs(sctx.Fields{
				"account_id": account.ID,
				"name":       account.Name,
			}).Info("Account recovered from rate limit")

			recoveredCount++
		}
	}

	return recoveredCount, nil
}
