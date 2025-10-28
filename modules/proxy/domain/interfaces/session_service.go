package interfaces

import (
	"context"
	"net/http"

	"claude-proxy/modules/proxy/domain/entities"
)

// SessionService defines the interface for session management operations
type SessionService interface {
	// CreateSession creates a new session and checks limits
	// Returns error if concurrent session limit is exceeded
	CreateSession(
		ctx context.Context,
		accountID string,
		tokenID string,
		req *http.Request,
	) (*entities.Session, error)

	// ValidateSession checks if a session is valid and within limits
	ValidateSession(
		ctx context.Context,
		sessionID string,
	) (bool, error)

	// RefreshSession extends session TTL
	RefreshSession(
		ctx context.Context,
		sessionID string,
	) error

	// RevokeSession manually revokes a session
	RevokeSession(
		ctx context.Context,
		sessionID string,
	) error

	// RevokeAccountSessions revokes all sessions for an account
	RevokeAccountSessions(
		ctx context.Context,
		accountID string,
	) (int, error)

	// GetAccountSessions retrieves all active sessions for an account
	GetAccountSessions(
		ctx context.Context,
		accountID string,
	) ([]*entities.Session, error)

	// GetAllSessions retrieves all active sessions (admin)
	GetAllSessions(ctx context.Context) ([]*entities.Session, error)

	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) (int, error)
}
