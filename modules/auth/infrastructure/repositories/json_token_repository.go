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

// JSONTokenRepository implements TokenPersistenceRepository using JSON file storage
// This repository ONLY handles disk I/O, no in-memory caching
type JSONTokenRepository struct {
	dataFolder string
	mu         sync.RWMutex // Only for file I/O concurrency control
}

// NewJSONTokenRepository creates a new JSON token repository
func NewJSONTokenRepository(dataFolder string) (interfaces.TokenPersistenceRepository, error) {
	repo := &JSONTokenRepository{
		dataFolder: expandPath(dataFolder),
	}

	// Create data folder if it doesn't exist
	if err := os.MkdirAll(repo.dataFolder, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create data folder: %w", err)
	}

	return repo, nil
}

// SaveAll persists all tokens to durable storage (batch operation)
func (r *JSONTokenRepository) SaveAll(ctx context.Context, tokens []*entities.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tokensFile := filepath.Join(r.dataFolder, "tokens.json")

	// Convert entities to DTOs
	dtos := make([]*dto.TokenPersistenceDTO, 0, len(tokens))
	for _, token := range tokens {
		dtos = append(dtos, dto.ToTokenPersistenceDTO(token))
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(dtos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Write to temporary file first (atomic write)
	tmpFile := tokensFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, tokensFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename tokens file: %w", err)
	}

	return nil
}

// LoadAll loads all tokens from durable storage
func (r *JSONTokenRepository) LoadAll(ctx context.Context) ([]*entities.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokensFile := filepath.Join(r.dataFolder, "tokens.json")

	data, err := os.ReadFile(tokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entities.Token{}, nil // No tokens yet
		}
		return nil, fmt.Errorf("failed to read tokens file: %w", err)
	}

	var dtos []*dto.TokenPersistenceDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return nil, fmt.Errorf("failed to parse tokens file: %w", err)
	}

	tokens := make([]*entities.Token, 0, len(dtos))
	for _, d := range dtos {
		tokens = append(tokens, dto.FromTokenPersistenceDTO(d))
	}

	return tokens, nil
}

// Create creates and persists a new token
func (r *JSONTokenRepository) Create(ctx context.Context, token *entities.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing tokens
	tokens, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Check for duplicates
	for _, t := range tokens {
		if t.ID == token.ID {
			return fmt.Errorf("token with ID already exists: %s", token.ID)
		}
	}

	// Add new token
	tokens = append(tokens, token)

	// Save all back to disk
	return r.saveToDisk(tokens)
}

// Update updates and persists an existing token
func (r *JSONTokenRepository) Update(ctx context.Context, token *entities.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing tokens
	tokens, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Find and update the token
	found := false
	for i, t := range tokens {
		if t.ID == token.ID {
			tokens[i] = token
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("token not found: %s", token.ID)
	}

	// Save all back to disk
	return r.saveToDisk(tokens)
}

// Delete deletes a token from persistent storage
func (r *JSONTokenRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing tokens
	tokens, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Find and remove the token
	found := false
	for i, t := range tokens {
		if t.ID == id {
			tokens = append(tokens[:i], tokens[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("token not found: %s", id)
	}

	// Save all back to disk
	return r.saveToDisk(tokens)
}

// loadFromDisk loads tokens from disk (internal helper, requires lock)
func (r *JSONTokenRepository) loadFromDisk() ([]*entities.Token, error) {
	tokensFile := filepath.Join(r.dataFolder, "tokens.json")

	data, err := os.ReadFile(tokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entities.Token{}, nil
		}
		return nil, fmt.Errorf("failed to read tokens file: %w", err)
	}

	var dtos []*dto.TokenPersistenceDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return nil, fmt.Errorf("failed to parse tokens file: %w", err)
	}

	tokens := make([]*entities.Token, 0, len(dtos))
	for _, d := range dtos {
		tokens = append(tokens, dto.FromTokenPersistenceDTO(d))
	}

	return tokens, nil
}

// saveToDisk saves tokens to disk (internal helper, requires lock)
func (r *JSONTokenRepository) saveToDisk(tokens []*entities.Token) error {
	tokensFile := filepath.Join(r.dataFolder, "tokens.json")

	// Convert entities to DTOs
	dtos := make([]*dto.TokenPersistenceDTO, 0, len(tokens))
	for _, token := range tokens {
		dtos = append(dtos, dto.ToTokenPersistenceDTO(token))
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(dtos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Write to temporary file first
	tmpFile := tokensFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, tokensFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename tokens file: %w", err)
	}

	return nil
}
