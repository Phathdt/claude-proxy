package interfaces

import (
	"context"

	"claude-proxy/modules/proxy/domain/entities"
)

// SessionRepository defines the interface for session persistence operations
type SessionRepository interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *entities.Session) error

	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, sessionID string) (*entities.Session, error)

	// ListSessionsByAccount retrieves all active sessions for an account
	ListSessionsByAccount(ctx context.Context, accountID string) ([]*entities.Session, error)

	// ListSessionsByToken retrieves all active sessions for a token
	ListSessionsByToken(ctx context.Context, tokenID string) ([]*entities.Session, error)

	// UpdateSession updates session metadata (last seen, expiry)
	UpdateSession(ctx context.Context, session *entities.Session) error

	// DeleteSession deletes a session by ID
	DeleteSession(ctx context.Context, sessionID string) error

	// DeleteSessionsByAccount deletes all sessions for an account
	DeleteSessionsByAccount(ctx context.Context, accountID string) error

	// CountActiveSessions counts active sessions for an account
	CountActiveSessions(ctx context.Context, accountID string) (int, error)

	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) (int, error)

	// ListAllSessions retrieves all active sessions (for admin)
	ListAllSessions(ctx context.Context) ([]*entities.Session, error)
}
