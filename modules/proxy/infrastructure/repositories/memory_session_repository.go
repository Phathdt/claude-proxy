package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"

	sctx "github.com/phathdt/service-context"
)

// MemorySessionRepository implements session repository with in-memory storage
type MemorySessionRepository struct {
	sessions map[string]*entities.Session // sessionID -> session
	accounts map[string][]string          // accountID -> []sessionID
	tokens   map[string][]string          // tokenID -> []sessionID
	mu       sync.RWMutex
	logger   sctx.Logger
}

// NewMemorySessionRepository creates a new in-memory session repository
func NewMemorySessionRepository(appLogger sctx.Logger) interfaces.SessionRepository {
	logger := appLogger.Withs(sctx.Fields{"component": "memory-session-repository"})

	return &MemorySessionRepository{
		sessions: make(map[string]*entities.Session),
		accounts: make(map[string][]string),
		tokens:   make(map[string][]string),
		logger:   logger,
	}
}

// CreateSession stores a new session in memory
func (r *MemorySessionRepository) CreateSession(ctx context.Context, session *entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store session
	r.sessions[session.ID] = session

	// Add to account index
	if _, exists := r.accounts[session.AccountID]; !exists {
		r.accounts[session.AccountID] = []string{}
	}
	r.accounts[session.AccountID] = append(r.accounts[session.AccountID], session.ID)

	// Add to token index
	if _, exists := r.tokens[session.TokenID]; !exists {
		r.tokens[session.TokenID] = []string{}
	}
	r.tokens[session.TokenID] = append(r.tokens[session.TokenID], session.ID)

	r.logger.Withs(sctx.Fields{
		"session_id": session.ID,
		"account_id": session.AccountID,
	}).Debug("Session created in memory")

	return nil
}

// GetSession retrieves a session by ID
func (r *MemorySessionRepository) GetSession(ctx context.Context, sessionID string) (*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// ListSessionsByAccount retrieves all sessions for an account
func (r *MemorySessionRepository) ListSessionsByAccount(
	ctx context.Context,
	accountID string,
) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionIDs, exists := r.accounts[accountID]
	if !exists {
		return []*entities.Session{}, nil
	}

	sessions := make([]*entities.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		if session, exists := r.sessions[sessionID]; exists {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// UpdateSession updates an existing session
func (r *MemorySessionRepository) UpdateSession(ctx context.Context, session *entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	r.sessions[session.ID] = session

	r.logger.Withs(sctx.Fields{"session_id": session.ID}).Debug("Session updated")
	return nil
}

// DeleteSession removes a session by ID
func (r *MemorySessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Remove from sessions map
	delete(r.sessions, sessionID)

	// Remove from account index
	r.accounts[session.AccountID] = r.removeFromSlice(r.accounts[session.AccountID], sessionID)
	if len(r.accounts[session.AccountID]) == 0 {
		delete(r.accounts, session.AccountID)
	}

	// Remove from token index
	r.tokens[session.TokenID] = r.removeFromSlice(r.tokens[session.TokenID], sessionID)
	if len(r.tokens[session.TokenID]) == 0 {
		delete(r.tokens, session.TokenID)
	}

	r.logger.Withs(sctx.Fields{"session_id": sessionID}).Debug("Session deleted")
	return nil
}

// DeleteSessionsByAccount removes all sessions for an account
func (r *MemorySessionRepository) DeleteSessionsByAccount(ctx context.Context, accountID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sessionIDs, exists := r.accounts[accountID]
	if !exists {
		return nil
	}

	// Remove all sessions
	for _, sessionID := range sessionIDs {
		if session, exists := r.sessions[sessionID]; exists {
			delete(r.sessions, sessionID)

			// Remove from token index
			r.tokens[session.TokenID] = r.removeFromSlice(r.tokens[session.TokenID], sessionID)
			if len(r.tokens[session.TokenID]) == 0 {
				delete(r.tokens, session.TokenID)
			}
		}
	}

	// Remove account index
	delete(r.accounts, accountID)

	r.logger.Withs(sctx.Fields{
		"account_id": accountID,
		"count":      len(sessionIDs),
	}).Debug("Account sessions deleted")

	return nil
}

// CountActiveSessions counts active (non-expired) sessions for an account
func (r *MemorySessionRepository) CountActiveSessions(ctx context.Context, accountID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionIDs, exists := r.accounts[accountID]
	if !exists {
		return 0, nil
	}

	count := 0
	now := time.Now()
	for _, sessionID := range sessionIDs {
		if session, exists := r.sessions[sessionID]; exists {
			if session.IsActive && now.Before(session.ExpiresAt) {
				count++
			}
		}
	}

	return count, nil
}

// CleanupExpiredSessions removes all expired sessions
func (r *MemorySessionRepository) CleanupExpiredSessions(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	expiredSessions := []string{}

	// Find expired sessions
	for sessionID, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	// Remove expired sessions
	for _, sessionID := range expiredSessions {
		session := r.sessions[sessionID]
		delete(r.sessions, sessionID)

		// Remove from account index
		r.accounts[session.AccountID] = r.removeFromSlice(r.accounts[session.AccountID], sessionID)
		if len(r.accounts[session.AccountID]) == 0 {
			delete(r.accounts, session.AccountID)
		}

		// Remove from token index
		r.tokens[session.TokenID] = r.removeFromSlice(r.tokens[session.TokenID], sessionID)
		if len(r.tokens[session.TokenID]) == 0 {
			delete(r.tokens, session.TokenID)
		}
	}

	if len(expiredSessions) > 0 {
		r.logger.Withs(sctx.Fields{"count": len(expiredSessions)}).Debug("Expired sessions cleaned up")
	}

	return len(expiredSessions), nil
}

// ListAllSessions retrieves all sessions
func (r *MemorySessionRepository) ListAllSessions(ctx context.Context) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*entities.Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// ListSessionsByToken retrieves all sessions for a token
func (r *MemorySessionRepository) ListSessionsByToken(
	ctx context.Context,
	tokenID string,
) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionIDs, exists := r.tokens[tokenID]
	if !exists {
		return []*entities.Session{}, nil
	}

	sessions := make([]*entities.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		if session, exists := r.sessions[sessionID]; exists {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// Helper function to remove an element from a slice
func (r *MemorySessionRepository) removeFromSlice(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
