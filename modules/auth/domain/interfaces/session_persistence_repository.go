package interfaces

import (
	"context"

	"claude-proxy/modules/auth/domain/entities"
)

// SessionPersistenceRepository defines the interface for durable session storage
// Implementation should prioritize data durability and persistence over speed
// All operations should persist to disk or permanent storage
type SessionPersistenceRepository interface {
	// SaveAll persists all sessions to durable storage (batch operation)
	SaveAll(ctx context.Context, sessions []*entities.Session) error

	// LoadAll loads all sessions from durable storage
	LoadAll(ctx context.Context) ([]*entities.Session, error)

	// CreateSession creates and persists a new session
	CreateSession(ctx context.Context, session *entities.Session) error

	// UpdateSession updates and persists an existing session
	UpdateSession(ctx context.Context, session *entities.Session) error

	// DeleteSession deletes a session from persistent storage
	DeleteSession(ctx context.Context, sessionID string) error
}
