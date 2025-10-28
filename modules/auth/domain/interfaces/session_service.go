package interfaces

import (
	"context"
	"net/http"

	"claude-proxy/modules/auth/domain/entities"
)

// SessionService defines the interface for session management operations
// Sessions track concurrent requests per client (IP + UserAgent)
type SessionService interface {
	// CreateSession creates a new session and checks global limits
	// Returns error if concurrent session limit is exceeded
	CreateSession(
		ctx context.Context,
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

	// GetAllSessions retrieves all active sessions (admin)
	GetAllSessions(ctx context.Context) ([]*entities.Session, error)

	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) (int, error)

	// Sync syncs in-memory data to persistent storage
	Sync(ctx context.Context) error

	// FinalSync performs final sync on graceful shutdown
	FinalSync(ctx context.Context) error
}
