package jobs

import (
	"context"
	"sync"
	"time"

	"claude-proxy/modules/proxy/domain/interfaces"

	sctx "github.com/phathdt/service-context"
	"github.com/robfig/cron/v3"
)

// Scheduler manages in-memory job scheduling with cron
type Scheduler struct {
	cron        *cron.Cron
	accountRepo interfaces.AccountRepository
	accountSvc  interfaces.AccountService
	logger      sctx.Logger
	mu          sync.Mutex
	running     bool
}

// NewScheduler creates a new in-memory scheduler
func NewScheduler(
	accountRepo interfaces.AccountRepository,
	accountSvc interfaces.AccountService,
	logger sctx.Logger,
) *Scheduler {
	return &Scheduler{
		cron:        cron.New(),
		accountRepo: accountRepo,
		accountSvc:  accountSvc,
		logger:      logger,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// Register the token refresh job to run every hour at minute 0
	_, err := s.cron.AddFunc("0 * * * *", s.RefreshTokensJob)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to register cron job")
		return err
	}

	s.cron.Start()
	s.running = true

	s.logger.Withs(sctx.Fields{
		"schedule": "0 * * * * (every hour at minute 0)",
	}).Info("In-memory job scheduler started")

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cron.Stop()
	s.running = false
	s.logger.Info("In-memory job scheduler stopped")
}

// RefreshTokensJob refreshes tokens for all active accounts
func (s *Scheduler) RefreshTokensJob() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	s.logger.Debug("Starting token refresh job")

	accounts, err := s.accountRepo.GetActiveAccounts(ctx)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to get active accounts for token refresh")
		return
	}

	if len(accounts) == 0 {
		s.logger.Debug("No active accounts to refresh")
		return
	}

	refreshedCount := 0
	failedCount := 0
	skippedCount := 0

	// Process each account
	for _, account := range accounts {
		// Check if token needs refresh (60s buffer)
		if !account.NeedsRefresh() {
			s.logger.Withs(sctx.Fields{
				"account_id": account.ID,
				"name":       account.Name,
			}).Debug("Token does not need refresh, skipping")
			skippedCount++
			continue
		}

		s.logger.Withs(sctx.Fields{
			"account_id": account.ID,
			"name":       account.Name,
		}).Debug("Refreshing account token")

		// Try to refresh the token
		_, err := s.accountSvc.GetValidToken(ctx, account.ID)
		if err != nil {
			s.logger.Withs(sctx.Fields{
				"account_id": account.ID,
				"name":       account.Name,
				"error":      err.Error(),
			}).Error("Failed to refresh account token")

			// Update account with error state
			account.UpdateRefreshError(err.Error())
			if err := s.accountRepo.Update(ctx, account); err != nil {
				s.logger.Withs(sctx.Fields{
					"account_id": account.ID,
					"error":      err.Error(),
				}).Error("Failed to update account with refresh error")
			}
			failedCount++
			continue
		}

		s.logger.Withs(sctx.Fields{
			"account_id": account.ID,
			"name":       account.Name,
		}).Info("Account token refreshed successfully")
		refreshedCount++
	}

	// Log summary
	s.logger.Withs(sctx.Fields{
		"total_accounts": len(accounts),
		"refreshed":      refreshedCount,
		"failed":         failedCount,
		"skipped":        skippedCount,
	}).Info("Token refresh job completed")
}
