package services

import (
	"context"
	"fmt"
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
		return err
	}

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
