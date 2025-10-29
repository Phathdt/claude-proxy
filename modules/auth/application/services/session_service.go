package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"claude-proxy/config"
	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"
	"claude-proxy/pkg/errors"

	"github.com/google/uuid"
	sctx "github.com/phathdt/service-context"
)

// SessionService implements session management with hybrid storage pattern
// Uses SessionCacheRepository for fast in-memory access and SessionPersistenceRepository for durability
type SessionService struct {
	cacheRepo       interfaces.SessionCacheRepository
	persistenceRepo interfaces.SessionPersistenceRepository
	maxConcurrent   int
	sessionTTL      time.Duration
	enabled         bool
	dirty           bool
	mu              sync.RWMutex
	logger          sctx.Logger
}

// NewSessionService creates a new session service with cache and persistence layers
func NewSessionService(
	cacheRepo interfaces.SessionCacheRepository,
	persistenceRepo interfaces.SessionPersistenceRepository,
	cfg *config.Config,
	appLogger sctx.Logger,
) interfaces.SessionService {
	logger := appLogger.Withs(sctx.Fields{"component": "session-service"})

	svc := &SessionService{
		cacheRepo:       cacheRepo,
		persistenceRepo: persistenceRepo,
		maxConcurrent:   cfg.Session.MaxConcurrent,
		sessionTTL:      cfg.Session.SessionTTL,
		enabled:         cfg.Session.Enabled,
		dirty:           false,
		logger:          logger,
	}

	// Load from persistent storage into cache on init
	if svc.enabled && persistenceRepo != nil {
		if err := svc.loadFromPersistence(); err != nil {
			logger.Withs(sctx.Fields{"error": err}).Warn("Failed to load sessions from persistence")
		}
	}

	return svc
}

// loadFromPersistence loads all sessions from persistent storage into cache
func (s *SessionService) loadFromPersistence() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessions, err := s.persistenceRepo.LoadAll(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load sessions from persistence: %w", err)
	}

	// Load each session into cache
	for _, session := range sessions {
		if err := s.cacheRepo.CreateSession(context.Background(), session); err != nil {
			s.logger.Withs(sctx.Fields{
				"session_id": session.ID,
				"error":      err,
			}).Warn("Failed to load session into cache")
		}
	}

	s.logger.Withs(sctx.Fields{"count": len(sessions)}).Info("Sessions loaded from persistence to cache")
	return nil
}

// markDirty marks data as changed
func (s *SessionService) markDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = true
}

// isDirty checks if data has changed
func (s *SessionService) isDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

// clearDirty clears the dirty flag
func (s *SessionService) clearDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = false
}

// Sync syncs cache data to persistent storage (called every 1 minute)
func (s *SessionService) Sync(ctx context.Context) error {
	if !s.enabled || s.persistenceRepo == nil {
		return nil
	}

	if !s.isDirty() {
		return nil // No changes, skip sync
	}

	s.logger.Debug("Syncing sessions to persistent storage")

	// Get all sessions from cache
	sessions, err := s.cacheRepo.ListAllSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions from cache: %w", err)
	}

	// Batch save all sessions to persistent storage
	if err := s.persistenceRepo.SaveAll(ctx, sessions); err != nil {
		s.logger.Withs(sctx.Fields{
			"error": err,
		}).Error("Failed to save sessions to persistence")
		return fmt.Errorf("failed to save sessions: %w", err)
	}

	s.clearDirty()
	s.logger.Withs(sctx.Fields{"count": len(sessions)}).Info("Sessions synced to persistent storage")
	return nil
}

// FinalSync performs final sync on graceful shutdown
func (s *SessionService) FinalSync(ctx context.Context) error {
	s.logger.Info("Performing final sync of sessions")
	return s.Sync(ctx)
}

// CreateSession creates a new session or reuses existing one (per client: IP + UserAgent)
func (s *SessionService) CreateSession(
	ctx context.Context,
	tokenID string,
	req *http.Request,
) (*entities.Session, error) {
	// If session limiting is disabled, skip
	if !s.enabled || s.cacheRepo == nil {
		return nil, nil
	}

	// Extract IP without port
	ipWithoutPort := s.getIPWithoutPort(req.RemoteAddr)
	userAgent := req.UserAgent()

	// Check if there's an existing active session for this IP + User-Agent
	existingSession := s.findExistingSession(ctx, ipWithoutPort, userAgent)
	if existingSession != nil {
		// Reuse existing session - just refresh it
		existingSession.Refresh(s.sessionTTL)
		if err := s.cacheRepo.UpdateSession(ctx, existingSession); err != nil {
			s.logger.Withs(sctx.Fields{"error": err}).Warn("Failed to refresh existing session")
		} else {
			s.markDirty()
			s.logger.Withs(sctx.Fields{
				"session_id": existingSession.ID,
				"ip_address": ipWithoutPort,
			}).Debug("Reused existing session")
		}
		return existingSession, nil
	}

	// No existing session found - check global active session count
	activeCount, err := s.cacheRepo.CountActiveSessions(ctx)
	if err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed to count active sessions")
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}

	// Check if global limit is exceeded
	if activeCount >= s.maxConcurrent {
		s.logger.Withs(sctx.Fields{
			"active_count":   activeCount,
			"max_concurrent": s.maxConcurrent,
		}).Warn("Global session limit exceeded")

		return nil, errors.NewRateLimitError(
			fmt.Sprintf("concurrent session limit exceeded: %d/%d active sessions", activeCount, s.maxConcurrent),
			map[string]interface{}{
				"active_count":   activeCount,
				"max_concurrent": s.maxConcurrent,
			},
		)
	}

	// Create new session
	now := time.Now()
	session := &entities.Session{
		ID:          uuid.Must(uuid.NewV7()).String(),
		TokenID:     tokenID,
		UserAgent:   userAgent,
		IPAddress:   ipWithoutPort, // Store IP without port for consistency
		CreatedAt:   now,
		LastSeenAt:  now,
		ExpiresAt:   now.Add(s.sessionTTL),
		IsActive:    true,
		RequestPath: req.URL.Path,
	}

	// Save to memory
	if err := s.cacheRepo.CreateSession(ctx, session); err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed to create session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{
		"session_id": session.ID,
		"token_id":   tokenID,
		"ip_address": session.IPAddress,
	}).Info("New session created")

	return session, nil
}

// getIPWithoutPort extracts IP address without port
func (s *SessionService) getIPWithoutPort(address string) string {
	// Handle IPv6 addresses like [::1]:12345 or IPv4 like 127.0.0.1:12345
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		// If SplitHostPort fails, assume it's already just an IP
		return address
	}
	return host
}

// findExistingSession looks for an active session with the same IP and User-Agent
func (s *SessionService) findExistingSession(
	ctx context.Context,
	ipWithoutPort, userAgent string,
) *entities.Session {
	sessions, err := s.cacheRepo.ListAllSessions(ctx)
	if err != nil {
		return nil
	}

	now := time.Now()
	for _, session := range sessions {
		// Match by IP (without port) and User-Agent
		sessionIP := s.getIPWithoutPort(session.IPAddress)
		if sessionIP == ipWithoutPort &&
			strings.EqualFold(session.UserAgent, userAgent) &&
			session.IsActive &&
			now.Before(session.ExpiresAt) {
			return session
		}
	}

	return nil
}

// ValidateSession checks if a session is valid and within limits
func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (bool, error) {
	if !s.enabled || s.cacheRepo == nil {
		return true, nil
	}

	session, err := s.cacheRepo.GetSession(ctx, sessionID)
	if err != nil {
		return false, err
	}

	if session.IsExpired() {
		return false, fmt.Errorf("session expired")
	}

	return session.IsActive, nil
}

// RefreshSession extends session TTL
func (s *SessionService) RefreshSession(ctx context.Context, sessionID string) error {
	if !s.enabled || s.cacheRepo == nil {
		return nil
	}

	session, err := s.cacheRepo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Refresh(s.sessionTTL)

	if err := s.cacheRepo.UpdateSession(ctx, session); err != nil {
		s.logger.Withs(sctx.Fields{"error": err, "session_id": sessionID}).Error("Failed to refresh session")
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"session_id": sessionID}).Debug("Session refreshed")
	return nil
}

// RevokeSession manually revokes a session
func (s *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	if !s.enabled || s.cacheRepo == nil {
		return fmt.Errorf("session limiting is not enabled")
	}

	if err := s.cacheRepo.DeleteSession(ctx, sessionID); err != nil {
		s.logger.Withs(sctx.Fields{"error": err, "session_id": sessionID}).Error("Failed to revoke session")
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"session_id": sessionID}).Info("Session revoked")
	return nil
}

// GetAllSessions retrieves all active sessions (admin)
func (s *SessionService) GetAllSessions(ctx context.Context) ([]*entities.Session, error) {
	if !s.enabled || s.cacheRepo == nil {
		return []*entities.Session{}, nil
	}
	return s.cacheRepo.ListAllSessions(ctx)
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) (int, error) {
	if !s.enabled || s.cacheRepo == nil {
		return 0, nil
	}

	count, err := s.cacheRepo.CleanupExpiredSessions(ctx)
	if err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed to cleanup expired sessions")
		return 0, err
	}

	if count > 0 {
		s.markDirty()
		s.logger.Withs(sctx.Fields{"cleaned_count": count}).Info("Expired sessions cleaned up")
	}

	return count, nil
}
