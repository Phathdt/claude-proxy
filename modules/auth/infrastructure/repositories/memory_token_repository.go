package repositories

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	sctx "github.com/phathdt/service-context"
)

// MemoryTokenRepository implements in-memory storage for tokens
type MemoryTokenRepository struct {
	tokens map[string]*entities.Token // tokenID -> token
	mu     sync.RWMutex
	logger sctx.Logger
}

// NewMemoryTokenRepository creates a new in-memory token repository
func NewMemoryTokenRepository(appLogger sctx.Logger) interfaces.TokenCacheRepository {
	logger := appLogger.Withs(sctx.Fields{"component": "memory-token-repository"})

	return &MemoryTokenRepository{
		tokens: make(map[string]*entities.Token),
		logger: logger,
	}
}

// Create creates a new token in memory
func (r *MemoryTokenRepository) Create(ctx context.Context, token *entities.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if token with same ID, key, or name already exists
	if _, exists := r.tokens[token.ID]; exists {
		return fmt.Errorf("token with ID already exists: %s", token.ID)
	}

	for _, t := range r.tokens {
		if t.Key == token.Key {
			return fmt.Errorf("token with key already exists")
		}
		if strings.EqualFold(t.Name, token.Name) {
			return fmt.Errorf("token with name already exists")
		}
	}

	r.tokens[token.ID] = token
	r.logger.Withs(sctx.Fields{"token_id": token.ID, "token_name": token.Name}).Debug("Token created in memory")
	return nil
}

// GetByID retrieves a token by ID
func (r *MemoryTokenRepository) GetByID(ctx context.Context, id string) (*entities.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, exists := r.tokens[id]
	if !exists {
		return nil, fmt.Errorf("token not found: %s", id)
	}

	return token, nil
}

// GetByKey retrieves a token by its key
func (r *MemoryTokenRepository) GetByKey(ctx context.Context, key string) (*entities.Token, error) {
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
func (r *MemoryTokenRepository) List(ctx context.Context) ([]*entities.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokens := make([]*entities.Token, 0, len(r.tokens))
	for _, token := range r.tokens {
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Update updates an existing token
func (r *MemoryTokenRepository) Update(ctx context.Context, token *entities.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tokens[token.ID]; !exists {
		return fmt.Errorf("token not found: %s", token.ID)
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
	r.logger.Withs(sctx.Fields{"token_id": token.ID}).Debug("Token updated in memory")
	return nil
}

// Delete removes a token by ID
func (r *MemoryTokenRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tokens[id]; !exists {
		return fmt.Errorf("token not found: %s", id)
	}

	delete(r.tokens, id)
	r.logger.Withs(sctx.Fields{"token_id": id}).Debug("Token deleted from memory")
	return nil
}
