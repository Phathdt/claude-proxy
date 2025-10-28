package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"claude-proxy/config"
	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"
	"claude-proxy/pkg/errors"

	"github.com/google/uuid"
	sctx "github.com/phathdt/service-context"
)

// SessionService implements session management business logic
type SessionService struct {
	sessionRepo   interfaces.SessionRepository
	maxConcurrent int
	sessionTTL    time.Duration
	enabled       bool
	logger        sctx.Logger
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionRepo interfaces.SessionRepository,
	cfg *config.Config,
	appLogger sctx.Logger,
) interfaces.SessionService {
	logger := appLogger.Withs(sctx.Fields{"component": "session-service"})

	return &SessionService{
		sessionRepo:   sessionRepo,
		maxConcurrent: cfg.Session.MaxConcurrent,
		sessionTTL:    cfg.Session.SessionTTL,
		enabled:       cfg.Session.Enabled,
		logger:        logger,
	}
}

// CreateSession creates a new session or reuses existing one
func (s *SessionService) CreateSession(
	ctx context.Context,
	accountID string,
	tokenID string,
	req *http.Request,
) (*entities.Session, error) {
	// If session limiting is disabled, skip
	if !s.enabled {
		return nil, nil
	}

	// Extract IP without port
	ipWithoutPort := s.getIPWithoutPort(req.RemoteAddr)
	userAgent := req.UserAgent()

	// Check if there's an existing active session for this IP + User-Agent + Account
	existingSession := s.findExistingSession(ctx, accountID, ipWithoutPort, userAgent)
	if existingSession != nil {
		// Reuse existing session - just refresh it
		existingSession.Refresh(s.sessionTTL)
		if err := s.sessionRepo.UpdateSession(ctx, existingSession); err != nil {
			s.logger.Withs(sctx.Fields{"error": err}).Warn("Failed to refresh existing session")
		} else {
			s.logger.Withs(sctx.Fields{
				"session_id": existingSession.ID,
				"account_id": accountID,
				"ip_address": ipWithoutPort,
			}).Debug("Reused existing session")
		}
		return existingSession, nil
	}

	// No existing session found - check current active session count
	activeCount, err := s.sessionRepo.CountActiveSessions(ctx, accountID)
	if err != nil {
		s.logger.Withs(sctx.Fields{"error": err, "account_id": accountID}).Error("Failed to count active sessions")
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}

	// Check if limit is exceeded
	if activeCount >= s.maxConcurrent {
		s.logger.Withs(sctx.Fields{
			"account_id":     accountID,
			"active_count":   activeCount,
			"max_concurrent": s.maxConcurrent,
		}).Warn("Session limit exceeded")

		return nil, errors.NewRateLimitError(
			fmt.Sprintf("concurrent session limit exceeded: %d/%d active sessions", activeCount, s.maxConcurrent),
			map[string]interface{}{
				"account_id":     accountID,
				"active_count":   activeCount,
				"max_concurrent": s.maxConcurrent,
			},
		)
	}

	// Create new session
	now := time.Now()
	session := &entities.Session{
		ID:          uuid.New().String(),
		AccountID:   accountID,
		TokenID:     tokenID,
		UserAgent:   userAgent,
		IPAddress:   ipWithoutPort, // Store IP without port for consistency
		CreatedAt:   now,
		LastSeenAt:  now,
		ExpiresAt:   now.Add(s.sessionTTL),
		IsActive:    true,
		RequestPath: req.URL.Path,
	}

	// Save to repository
	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed to create session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Withs(sctx.Fields{
		"session_id": session.ID,
		"account_id": accountID,
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
	accountID, ipWithoutPort, userAgent string,
) *entities.Session {
	sessions, err := s.sessionRepo.ListSessionsByAccount(ctx, accountID)
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
	if !s.enabled {
		return true, nil
	}

	session, err := s.sessionRepo.GetSession(ctx, sessionID)
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
	if !s.enabled {
		return nil
	}

	session, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Refresh(s.sessionTTL)

	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		s.logger.Withs(sctx.Fields{"error": err, "session_id": sessionID}).Error("Failed to refresh session")
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	s.logger.Withs(sctx.Fields{"session_id": sessionID}).Debug("Session refreshed")
	return nil
}

// RevokeSession manually revokes a session
func (s *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	if !s.enabled || s.sessionRepo == nil {
		return fmt.Errorf("session limiting is not enabled")
	}

	if err := s.sessionRepo.DeleteSession(ctx, sessionID); err != nil {
		s.logger.Withs(sctx.Fields{"error": err, "session_id": sessionID}).Error("Failed to revoke session")
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	s.logger.Withs(sctx.Fields{"session_id": sessionID}).Info("Session revoked")
	return nil
}

// RevokeAccountSessions revokes all sessions for an account
func (s *SessionService) RevokeAccountSessions(ctx context.Context, accountID string) (int, error) {
	if !s.enabled || s.sessionRepo == nil {
		return 0, fmt.Errorf("session limiting is not enabled")
	}

	sessions, err := s.sessionRepo.ListSessionsByAccount(ctx, accountID)
	if err != nil {
		return 0, err
	}

	count := len(sessions)

	if err := s.sessionRepo.DeleteSessionsByAccount(ctx, accountID); err != nil {
		s.logger.Withs(sctx.Fields{"error": err, "account_id": accountID}).Error("Failed to revoke account sessions")
		return 0, fmt.Errorf("failed to revoke account sessions: %w", err)
	}

	s.logger.Withs(sctx.Fields{
		"account_id":    accountID,
		"revoked_count": count,
	}).Info("Account sessions revoked")

	return count, nil
}

// GetAccountSessions retrieves all active sessions for an account
func (s *SessionService) GetAccountSessions(ctx context.Context, accountID string) ([]*entities.Session, error) {
	if !s.enabled || s.sessionRepo == nil {
		return []*entities.Session{}, nil
	}
	return s.sessionRepo.ListSessionsByAccount(ctx, accountID)
}

// GetAllSessions retrieves all active sessions (admin)
func (s *SessionService) GetAllSessions(ctx context.Context) ([]*entities.Session, error) {
	if !s.enabled || s.sessionRepo == nil {
		return []*entities.Session{}, nil
	}
	return s.sessionRepo.ListAllSessions(ctx)
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) (int, error) {
	if !s.enabled {
		return 0, nil
	}

	count, err := s.sessionRepo.CleanupExpiredSessions(ctx)
	if err != nil {
		s.logger.Withs(sctx.Fields{"error": err}).Error("Failed to cleanup expired sessions")
		return 0, err
	}

	if count > 0 {
		s.logger.Withs(sctx.Fields{"cleaned_count": count}).Info("Expired sessions cleaned up")
	}

	return count, nil
}
