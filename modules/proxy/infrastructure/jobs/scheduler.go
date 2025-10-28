package jobs

import (
	"context"
	"sync"
	"time"

	"claude-proxy/modules/auth/domain/interfaces"

	sctx "github.com/phathdt/service-context"
	"github.com/robfig/cron/v3"
)

// Scheduler manages in-memory job scheduling with cron
type Scheduler struct {
	cron       *cron.Cron
	accountSvc interfaces.AccountService
	logger     sctx.Logger
	mu         sync.Mutex
	running    bool
}

// NewScheduler creates a new in-memory scheduler
func NewScheduler(
	accountSvc interfaces.AccountService,
	logger sctx.Logger,
) *Scheduler {
	return &Scheduler{
		cron:       cron.New(),
		accountSvc: accountSvc,
		logger:     logger,
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

// RefreshTokensJob refreshes tokens for all active accounts and recovers rate-limited accounts
func (s *Scheduler) RefreshTokensJob() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	s.logger.Debug("Starting token refresh and recovery job")

	// Step 1: Recover rate-limited accounts with expired limits
	recoveredCount, err := s.accountSvc.RecoverRateLimitedAccounts(ctx)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to recover rate limited accounts")
	} else if recoveredCount > 0 {
		s.logger.Withs(sctx.Fields{
			"recovered_count": recoveredCount,
		}).Info("Recovered rate limited accounts")
	}

	// Step 2: Refresh active accounts using service method
	refreshedCount, failedCount, skippedCount, err := s.accountSvc.RefreshAllAccounts(ctx)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to refresh accounts")
		return
	}

	// Log summary
	s.logger.Withs(sctx.Fields{
		"refreshed": refreshedCount,
		"failed":    failedCount,
		"skipped":   skippedCount,
		"recovered": recoveredCount,
	}).Info("Token refresh and recovery job completed")
}
