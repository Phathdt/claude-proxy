package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// SessionRepository defines the interface for session persistence operations
type SessionRepository interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *entities.Session) error

	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, sessionID string) (*entities.Session, error)

	// ListSessionsByToken retrieves all active sessions for a token
	ListSessionsByToken(ctx context.Context, tokenID string) ([]*entities.Session, error)

	// UpdateSession updates session metadata (last seen, expiry)
	UpdateSession(ctx context.Context, session *entities.Session) error

	// DeleteSession deletes a session by ID
	DeleteSession(ctx context.Context, sessionID string) error

	// CountActiveSessions counts total active sessions globally
	CountActiveSessions(ctx context.Context) (int, error)

	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) (int, error)

	// ListAllSessions retrieves all active sessions (for admin)
	ListAllSessions(ctx context.Context) ([]*entities.Session, error)
}
