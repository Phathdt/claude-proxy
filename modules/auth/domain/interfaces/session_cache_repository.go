package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// SessionCacheRepository defines the interface for fast, volatile session storage
// Implementation should prioritize speed over durability
type SessionCacheRepository interface {
	// CreateSession creates a new session in cache
	CreateSession(ctx context.Context, session *entities.Session) error

	// GetSession retrieves a session by ID from cache
	GetSession(ctx context.Context, sessionID string) (*entities.Session, error)

	// ListSessionsByToken retrieves all active sessions for a token from cache
	ListSessionsByToken(ctx context.Context, tokenID string) ([]*entities.Session, error)

	// UpdateSession updates session metadata in cache
	UpdateSession(ctx context.Context, session *entities.Session) error

	// DeleteSession deletes a session by ID from cache
	DeleteSession(ctx context.Context, sessionID string) error

	// CountActiveSessions counts total active sessions globally from cache
	CountActiveSessions(ctx context.Context) (int, error)

	// CleanupExpiredSessions removes expired sessions from cache
	CleanupExpiredSessions(ctx context.Context) (int, error)

	// ListAllSessions retrieves all sessions from cache (for admin)
	ListAllSessions(ctx context.Context) ([]*entities.Session, error)
}
