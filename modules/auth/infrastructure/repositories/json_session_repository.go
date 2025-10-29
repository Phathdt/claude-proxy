package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"
)

// JSONSessionRepository implements SessionPersistenceRepository using JSON file storage
// This repository ONLY handles disk I/O, no in-memory caching
type JSONSessionRepository struct {
	dataFolder string
	mu         sync.RWMutex // Only for file I/O concurrency control
}

// NewJSONSessionRepository creates a new JSON session repository
func NewJSONSessionRepository(dataFolder string) (interfaces.SessionPersistenceRepository, error) {
	repo := &JSONSessionRepository{
		dataFolder: expandPath(dataFolder),
	}

	// Create data folder if it doesn't exist
	if err := os.MkdirAll(repo.dataFolder, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create data folder: %w", err)
	}

	return repo, nil
}

// SaveAll persists all sessions to durable storage (batch operation)
func (r *JSONSessionRepository) SaveAll(ctx context.Context, sessions []*entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sessionsFile := filepath.Join(r.dataFolder, "sessions.json")

	// Convert entities to DTOs
	dtos := make([]*dto.SessionPersistenceDTO, 0, len(sessions))
	for _, session := range sessions {
		dtos = append(dtos, dto.ToSessionPersistenceDTO(session))
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(dtos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	// Write to temporary file first (atomic write)
	tmpFile := sessionsFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write sessions file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, sessionsFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename sessions file: %w", err)
	}

	return nil
}

// LoadAll loads all sessions from durable storage
func (r *JSONSessionRepository) LoadAll(ctx context.Context) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionsFile := filepath.Join(r.dataFolder, "sessions.json")

	data, err := os.ReadFile(sessionsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entities.Session{}, nil // No sessions yet
		}
		return nil, fmt.Errorf("failed to read sessions file: %w", err)
	}

	var dtos []*dto.SessionPersistenceDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return nil, fmt.Errorf("failed to parse sessions file: %w", err)
	}

	sessions := make([]*entities.Session, 0, len(dtos))
	for _, d := range dtos {
		sessions = append(sessions, dto.FromSessionPersistenceDTO(d))
	}

	return sessions, nil
}

// CreateSession creates and persists a new session
func (r *JSONSessionRepository) CreateSession(ctx context.Context, session *entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing sessions
	sessions, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Check for duplicates
	for _, s := range sessions {
		if s.ID == session.ID {
			return fmt.Errorf("session with ID already exists: %s", session.ID)
		}
	}

	// Add new session
	sessions = append(sessions, session)

	// Save all back to disk
	return r.saveToDisk(sessions)
}

// UpdateSession updates and persists an existing session
func (r *JSONSessionRepository) UpdateSession(ctx context.Context, session *entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing sessions
	sessions, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Find and update the session
	found := false
	for i, s := range sessions {
		if s.ID == session.ID {
			sessions[i] = session
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	// Save all back to disk
	return r.saveToDisk(sessions)
}

// DeleteSession deletes a session from persistent storage
func (r *JSONSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing sessions
	sessions, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Find and remove the session
	found := false
	for i, s := range sessions {
		if s.ID == sessionID {
			sessions = append(sessions[:i], sessions[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Save all back to disk
	return r.saveToDisk(sessions)
}

// loadFromDisk loads sessions from disk (internal helper, requires lock)
func (r *JSONSessionRepository) loadFromDisk() ([]*entities.Session, error) {
	sessionsFile := filepath.Join(r.dataFolder, "sessions.json")

	data, err := os.ReadFile(sessionsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entities.Session{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions file: %w", err)
	}

	var dtos []*dto.SessionPersistenceDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return nil, fmt.Errorf("failed to parse sessions file: %w", err)
	}

	sessions := make([]*entities.Session, 0, len(dtos))
	for _, d := range dtos {
		sessions = append(sessions, dto.FromSessionPersistenceDTO(d))
	}

	return sessions, nil
}

// saveToDisk saves sessions to disk (internal helper, requires lock)
func (r *JSONSessionRepository) saveToDisk(sessions []*entities.Session) error {
	sessionsFile := filepath.Join(r.dataFolder, "sessions.json")

	// Convert entities to DTOs
	dtos := make([]*dto.SessionPersistenceDTO, 0, len(sessions))
	for _, session := range sessions {
		dtos = append(dtos, dto.ToSessionPersistenceDTO(session))
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(dtos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	// Write to temporary file first
	tmpFile := sessionsFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write sessions file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, sessionsFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename sessions file: %w", err)
	}

	return nil
}
