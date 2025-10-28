package jobs

import (
	"context"
	"sync"
	"time"

	"claude-proxy/config"
	"claude-proxy/modules/auth/domain/interfaces"

	sctx "github.com/phathdt/service-context"
	"github.com/robfig/cron/v3"
)

// SessionCleanupScheduler handles periodic cleanup of expired sessions
type SessionCleanupScheduler struct {
	sessionService interfaces.SessionService
	interval       time.Duration
	cron           *cron.Cron
	mu             sync.Mutex
	logger         sctx.Logger
}

// NewSessionCleanupScheduler creates a new session cleanup scheduler
func NewSessionCleanupScheduler(
	sessionService interfaces.SessionService,
	cfg *config.Config,
	appLogger sctx.Logger,
) *SessionCleanupScheduler {
	logger := appLogger.Withs(sctx.Fields{"component": "session-cleanup-scheduler"})

	return &SessionCleanupScheduler{
		sessionService: sessionService,
		interval:       cfg.Session.CleanupInterval,
		cron:           cron.New(),
		logger:         logger,
	}
}

// Start starts the cleanup scheduler
func (s *SessionCleanupScheduler) Start() error {
	s.logger.Withs(sctx.Fields{
		"interval": s.interval.String(),
	}).Info("Starting session cleanup scheduler")

	// Convert interval to cron expression
	// For simplicity, use @every syntax
	cronExpr := "@every " + s.interval.String()

	_, err := s.cron.AddFunc(cronExpr, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.runCleanup()
	})
	if err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed to schedule cleanup job")
		return err
	}

	s.cron.Start()
	s.logger.Info("Session cleanup scheduler started")

	return nil
}

// Stop stops the cleanup scheduler
func (s *SessionCleanupScheduler) Stop() {
	s.logger.Info("Stopping session cleanup scheduler")
	s.cron.Stop()
}

// runCleanup executes the cleanup job
func (s *SessionCleanupScheduler) runCleanup() {
	start := time.Now()
	s.logger.Debug("Running session cleanup job")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	count, err := s.sessionService.CleanupExpiredSessions(ctx)
	if err != nil {
		s.logger.Withs(sctx.Fields{
			"error":    err.Error(),
			"duration": time.Since(start).String(),
		}).Error("Session cleanup job failed")
		return
	}

	if count > 0 {
		s.logger.Withs(sctx.Fields{
			"cleaned_count": count,
			"duration":      time.Since(start).String(),
		}).Info("Session cleanup job completed")
	} else {
		s.logger.Withs(sctx.Fields{
			"duration": time.Since(start).String(),
		}).Debug("Session cleanup job completed (no sessions to clean)")
	}
}
