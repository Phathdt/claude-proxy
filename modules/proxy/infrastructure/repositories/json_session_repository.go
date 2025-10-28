package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"

	sctx "github.com/phathdt/service-context"
)

// JSONSessionRepository implements session repository with JSON file persistence
type JSONSessionRepository struct {
	filePath string
	sessions map[string]*entities.Session // sessionID -> session
	accounts map[string][]string          // accountID -> []sessionID
	tokens   map[string][]string          // tokenID -> []sessionID
	mu       sync.RWMutex
	logger   sctx.Logger
}

// NewJSONSessionRepository creates a new JSON-based session repository
func NewJSONSessionRepository(dataFolder string, appLogger sctx.Logger) (interfaces.SessionRepository, error) {
	logger := appLogger.Withs(sctx.Fields{"component": "json-session-repository"})

	// Ensure data folder exists
	if err := os.MkdirAll(dataFolder, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data folder: %w", err)
	}

	filePath := filepath.Join(dataFolder, "sessions.json")

	repo := &JSONSessionRepository{
		filePath: filePath,
		sessions: make(map[string]*entities.Session),
		accounts: make(map[string][]string),
		tokens:   make(map[string][]string),
		logger:   logger,
	}

	// Load existing sessions from file
	if err := repo.load(); err != nil {
		logger.Withs(sctx.Fields{"error": err}).Warn("Failed to load sessions from file, starting fresh")
	}

	return repo, nil
}

// load reads sessions from JSON file
func (r *JSONSessionRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return nil // No file yet, start with empty
	}

	// Read file
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to read sessions file: %w", err)
	}

	// Parse JSON
	var sessions []*entities.Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return fmt.Errorf("failed to unmarshal sessions: %w", err)
	}

	// Build indexes
	for _, session := range sessions {
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
	}

	r.logger.Withs(sctx.Fields{"count": len(sessions)}).Info("Sessions loaded from file")
	return nil
}

// save writes sessions to JSON file
func (r *JSONSessionRepository) save() error {
	// Convert map to slice
	sessions := make([]*entities.Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	// Write to file
	if err := os.WriteFile(r.filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write sessions file: %w", err)
	}

	return nil
}

// CreateSession stores a new session in JSON file
func (r *JSONSessionRepository) CreateSession(ctx context.Context, session *entities.Session) error {
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

	// Persist to file
	if err := r.save(); err != nil {
		r.logger.Withs(sctx.Fields{"error": err}).Error("Failed to save sessions to file")
		return err
	}

	r.logger.Withs(sctx.Fields{
		"session_id": session.ID,
		"account_id": session.AccountID,
	}).Debug("Session created and persisted")

	return nil
}

// GetSession retrieves a session by ID
func (r *JSONSessionRepository) GetSession(ctx context.Context, sessionID string) (*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// ListSessionsByAccount retrieves all sessions for an account
func (r *JSONSessionRepository) ListSessionsByAccount(
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

// ListSessionsByToken retrieves all sessions for a token
func (r *JSONSessionRepository) ListSessionsByToken(ctx context.Context, tokenID string) ([]*entities.Session, error) {
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

// UpdateSession updates an existing session
func (r *JSONSessionRepository) UpdateSession(ctx context.Context, session *entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	r.sessions[session.ID] = session

	// Persist to file
	if err := r.save(); err != nil {
		r.logger.Withs(sctx.Fields{"error": err}).Error("Failed to save sessions to file")
		return err
	}

	r.logger.Withs(sctx.Fields{"session_id": session.ID}).Debug("Session updated")
	return nil
}

// DeleteSession removes a session by ID
func (r *JSONSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
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

	// Persist to file
	if err := r.save(); err != nil {
		r.logger.Withs(sctx.Fields{"error": err}).Error("Failed to save sessions to file")
		return err
	}

	r.logger.Withs(sctx.Fields{"session_id": sessionID}).Debug("Session deleted")
	return nil
}

// DeleteSessionsByAccount removes all sessions for an account
func (r *JSONSessionRepository) DeleteSessionsByAccount(ctx context.Context, accountID string) error {
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

	// Persist to file
	if err := r.save(); err != nil {
		r.logger.Withs(sctx.Fields{"error": err}).Error("Failed to save sessions to file")
		return err
	}

	r.logger.Withs(sctx.Fields{
		"account_id": accountID,
		"count":      len(sessionIDs),
	}).Debug("Account sessions deleted")

	return nil
}

// CountActiveSessions counts active (non-expired) sessions for an account
func (r *JSONSessionRepository) CountActiveSessions(ctx context.Context, accountID string) (int, error) {
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
func (r *JSONSessionRepository) CleanupExpiredSessions(ctx context.Context) (int, error) {
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

	if len(expiredSessions) == 0 {
		return 0, nil
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

	// Persist to file
	if err := r.save(); err != nil {
		r.logger.Withs(sctx.Fields{"error": err}).Error("Failed to save sessions to file")
		return 0, err
	}

	r.logger.Withs(sctx.Fields{"count": len(expiredSessions)}).Debug("Expired sessions cleaned up")
	return len(expiredSessions), nil
}

// ListAllSessions retrieves all sessions
func (r *JSONSessionRepository) ListAllSessions(ctx context.Context) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*entities.Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Helper function to remove an element from a slice
func (r *JSONSessionRepository) removeFromSlice(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
