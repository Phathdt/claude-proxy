package jobs

import (
	"context"
	"sync"
	"time"

	"claude-proxy/modules/auth/domain/interfaces"

	sctx "github.com/phathdt/service-context"
	"github.com/robfig/cron/v3"
)

// SyncScheduler handles periodic sync of in-memory data to persistent storage
type SyncScheduler struct {
	accountService interfaces.AccountService
	tokenService   interfaces.TokenService
	sessionService interfaces.SessionService
	interval       time.Duration
	cron           *cron.Cron
	mu             sync.Mutex
	logger         sctx.Logger
}

// NewSyncScheduler creates a new sync scheduler
func NewSyncScheduler(
	accountService interfaces.AccountService,
	tokenService interfaces.TokenService,
	sessionService interfaces.SessionService,
	syncInterval time.Duration,
	appLogger sctx.Logger,
) *SyncScheduler {
	logger := appLogger.Withs(sctx.Fields{"component": "sync-scheduler"})

	return &SyncScheduler{
		accountService: accountService,
		tokenService:   tokenService,
		sessionService: sessionService,
		interval:       syncInterval,
		cron:           cron.New(),
		logger:         logger,
	}
}

// Start starts the sync scheduler
func (s *SyncScheduler) Start() error {
	s.logger.Withs(sctx.Fields{
		"interval": s.interval.String(),
	}).Info("Starting sync scheduler")

	// Convert interval to cron expression
	// Use cron syntax to run at exact minute boundaries (at :00 seconds)
	// Examples:
	//   1 minute  -> "* * * * *"   (every minute)
	//   5 minutes -> "*/5 * * * *" (every 5 minutes)
	//   Other     -> "@every Xm"   (interval-based fallback)
	var cronExpr string
	if s.interval == 1*time.Minute {
		cronExpr = "* * * * *" // Every minute at :00 seconds
	} else if s.interval == 5*time.Minute {
		cronExpr = "*/5 * * * *" // Every 5 minutes at :00 seconds
	} else if s.interval == 10*time.Minute {
		cronExpr = "*/10 * * * *" // Every 10 minutes at :00 seconds
	} else {
		cronExpr = "@every " + s.interval.String() // Fallback for custom intervals
	}

	_, err := s.cron.AddFunc(cronExpr, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.runSync()
	})
	if err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed to schedule sync job")
		return err
	}

	s.cron.Start()
	s.logger.Withs(sctx.Fields{
		"schedule": cronExpr,
	}).Info("Sync scheduler started")

	return nil
}

// Stop stops the sync scheduler
func (s *SyncScheduler) Stop() {
	s.logger.Info("Stopping sync scheduler")
	s.cron.Stop()
}

// runSync executes the sync job
func (s *SyncScheduler) runSync() {
	start := time.Now()
	s.logger.Debug("Running sync job")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Sync accounts
	if err := s.accountService.Sync(ctx); err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to sync accounts")
	}

	// Sync tokens
	if err := s.tokenService.Sync(ctx); err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to sync tokens")
	}

	// Sync sessions
	if err := s.sessionService.Sync(ctx); err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to sync sessions")
	}

	s.logger.Withs(sctx.Fields{
		"duration": time.Since(start).String(),
	}).Debug("Sync job completed")
}

// FinalSync performs final sync before shutdown
func (s *SyncScheduler) FinalSync() error {
	s.logger.Info("Performing final sync before shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Final sync for all services
	if err := s.accountService.FinalSync(ctx); err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed final sync of accounts")
		return err
	}

	if err := s.tokenService.FinalSync(ctx); err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed final sync of tokens")
		return err
	}

	if err := s.sessionService.FinalSync(ctx); err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed final sync of sessions")
		return err
	}

	s.logger.Info("Final sync completed successfully")
	return nil
}
