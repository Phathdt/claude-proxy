package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	"github.com/google/uuid"
)

// TokenDTO represents the JSON structure for token persistence
type TokenDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	Status     string `json:"status"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	UsageCount int    `json:"usage_count"`
	LastUsedAt *int64 `json:"last_used_at,omitempty"`
}

// JSONTokenRepository implements TokenRepository using JSON file storage
type JSONTokenRepository struct {
	dataFolder string
	tokens     map[string]*entities.Token
	mu         sync.RWMutex
}

// NewJSONTokenRepository creates a new JSON token repository
func NewJSONTokenRepository(dataFolder string) (interfaces.TokenRepository, error) {
	repo := &JSONTokenRepository{
		dataFolder: expandPath(dataFolder),
		tokens:     make(map[string]*entities.Token),
	}

	// Create data folder if it doesn't exist
	if err := os.MkdirAll(repo.dataFolder, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create data folder: %w", err)
	}

	// Load existing tokens
	if err := repo.load(); err != nil {
		return nil, fmt.Errorf("failed to load tokens: %w", err)
	}

	return repo, nil
}

// Create creates a new token
func (r *JSONTokenRepository) Create(ctx context.Context, token *entities.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if token with same key or name already exists
	for _, t := range r.tokens {
		if t.Key == token.Key {
			return fmt.Errorf("token with key already exists")
		}
		if strings.EqualFold(t.Name, token.Name) {
			return fmt.Errorf("token with name already exists")
		}
	}

	// Generate ID if not set (should never happen, but defensive)
	if token.ID == "" {
		token.ID = uuid.Must(uuid.NewV7()).String()
	}

	r.tokens[token.ID] = token
	return r.save()
}

// GetByID retrieves a token by ID
func (r *JSONTokenRepository) GetByID(ctx context.Context, id string) (*entities.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, exists := r.tokens[id]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	return token, nil
}

// GetByKey retrieves a token by its key
func (r *JSONTokenRepository) GetByKey(ctx context.Context, key string) (*entities.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, token := range r.tokens {
		if token.Key == key {
			return token, nil
		}
	}

	return nil, fmt.Errorf("token not found")
}

// List retrieves all tokens
func (r *JSONTokenRepository) List(ctx context.Context) ([]*entities.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokens := make([]*entities.Token, 0, len(r.tokens))
	for _, token := range r.tokens {
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// Update updates an existing token
func (r *JSONTokenRepository) Update(ctx context.Context, token *entities.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tokens[token.ID]; !exists {
		return fmt.Errorf("token not found")
	}

	// Check if key or name changed and conflicts with another token
	for id, t := range r.tokens {
		if id != token.ID {
			if t.Key == token.Key {
				return fmt.Errorf("token with key already exists")
			}
			if strings.EqualFold(t.Name, token.Name) {
				return fmt.Errorf("token with name already exists")
			}
		}
	}

	r.tokens[token.ID] = token
	return r.save()
}

// Delete deletes a token by ID
func (r *JSONTokenRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tokens[id]; !exists {
		return fmt.Errorf("token not found")
	}

	delete(r.tokens, id)
	return r.save()
}

// load loads tokens from disk
func (r *JSONTokenRepository) load() error {
	tokensFile := filepath.Join(r.dataFolder, "tokens.json")

	data, err := os.ReadFile(tokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No tokens yet
		}
		return fmt.Errorf("failed to read tokens file: %w", err)
	}

	var dtos []*TokenDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return fmt.Errorf("failed to parse tokens file: %w", err)
	}

	for _, dto := range dtos {
		token := dtoToEntity(dto)
		r.tokens[token.ID] = token
	}

	return nil
}

// save persists tokens to disk
func (r *JSONTokenRepository) save() error {
	tokensFile := filepath.Join(r.dataFolder, "tokens.json")

	tokens := make([]*TokenDTO, 0, len(r.tokens))
	for _, token := range r.tokens {
		tokens = append(tokens, entityToDTO(token))
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	if err := os.WriteFile(tokensFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	return nil
}

// entityToDTO converts entity to DTO
func entityToDTO(token *entities.Token) *TokenDTO {
	dto := &TokenDTO{
		ID:         token.ID,
		Name:       token.Name,
		Key:        token.Key,
		Status:     string(token.Status),
		CreatedAt:  token.CreatedAt.Unix(),
		UpdatedAt:  token.UpdatedAt.Unix(),
		UsageCount: token.UsageCount,
	}

	if token.LastUsedAt != nil {
		lastUsed := token.LastUsedAt.Unix()
		dto.LastUsedAt = &lastUsed
	}

	return dto
}

// dtoToEntity converts DTO to entity
func dtoToEntity(dto *TokenDTO) *entities.Token {
	token := &entities.Token{
		ID:         dto.ID,
		Name:       dto.Name,
		Key:        dto.Key,
		Status:     entities.TokenStatus(dto.Status),
		CreatedAt:  time.Unix(dto.CreatedAt, 0),
		UpdatedAt:  time.Unix(dto.UpdatedAt, 0),
		UsageCount: dto.UsageCount,
	}

	if dto.LastUsedAt != nil {
		lastUsed := time.Unix(*dto.LastUsedAt, 0)
		token.LastUsedAt = &lastUsed
	}

	return token
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
