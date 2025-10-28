package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	"github.com/google/uuid"
	sctx "github.com/phathdt/service-context"
	"github.com/phathdt/service-context/core"
)

// TokenService implements token management with hybrid storage
type TokenService struct {
	memoryRepo interfaces.TokenRepository
	jsonRepo   interfaces.TokenRepository
	dirty      bool
	mu         sync.RWMutex
	logger     sctx.Logger
}

// NewTokenService creates a new token service
func NewTokenService(
	memoryRepo interfaces.TokenRepository,
	jsonRepo interfaces.TokenRepository,
	appLogger sctx.Logger,
) interfaces.TokenService {
	logger := appLogger.Withs(sctx.Fields{"component": "token-service"})

	svc := &TokenService{
		memoryRepo: memoryRepo,
		jsonRepo:   jsonRepo,
		dirty:      false,
		logger:     logger,
	}

	// Load from JSON into memory on init
	if err := svc.loadFromJSON(); err != nil {
		logger.Withs(sctx.Fields{"error": err}).Warn("Failed to load tokens from JSON")
	}

	return svc
}

// loadFromJSON loads all tokens from JSON into memory
func (s *TokenService) loadFromJSON() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.jsonRepo.List(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list tokens from JSON: %w", err)
	}

	// Load each token into memory
	for _, token := range tokens {
		if err := s.memoryRepo.Create(context.Background(), token); err != nil {
			s.logger.Withs(sctx.Fields{
				"token_id": token.ID,
				"error":    err,
			}).Warn("Failed to load token into memory")
		}
	}

	s.logger.Withs(sctx.Fields{"count": len(tokens)}).Info("Tokens loaded from JSON to memory")
	return nil
}

// markDirty marks data as changed
func (s *TokenService) markDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = true
}

// isDirty checks if data has changed
func (s *TokenService) isDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

// clearDirty clears the dirty flag
func (s *TokenService) clearDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = false
}

// Sync syncs in-memory data to JSON (called every 1 minute)
func (s *TokenService) Sync(ctx context.Context) error {
	if !s.isDirty() {
		return nil // No changes, skip sync
	}

	s.logger.Debug("Syncing tokens to JSON")

	// Get all tokens from memory
	tokens, err := s.memoryRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tokens from memory: %w", err)
	}

	// Sync each token to JSON
	for _, token := range tokens {
		// Check if exists in JSON
		existing, err := s.jsonRepo.GetByID(ctx, token.ID)
		if err != nil {
			// Doesn't exist, create
			if err := s.jsonRepo.Create(ctx, token); err != nil {
				s.logger.Withs(sctx.Fields{
					"token_id": token.ID,
					"error":    err,
				}).Error("Failed to create token in JSON")
				continue
			}
		} else if existing != nil {
			// Exists, update
			if err := s.jsonRepo.Update(ctx, token); err != nil {
				s.logger.Withs(sctx.Fields{
					"token_id": token.ID,
					"error":    err,
				}).Error("Failed to update token in JSON")
				continue
			}
		}
	}

	s.clearDirty()
	s.logger.Withs(sctx.Fields{"count": len(tokens)}).Info("Tokens synced to JSON")
	return nil
}

// FinalSync performs final sync on graceful shutdown
func (s *TokenService) FinalSync(ctx context.Context) error {
	s.logger.Info("Performing final sync of tokens")
	return s.Sync(ctx)
}

// CreateToken creates a new token
func (s *TokenService) CreateToken(
	ctx context.Context,
	name, key string,
	status entities.TokenStatus,
	role entities.TokenRole,
) (*entities.Token, error) {
	// Default to user role if not specified
	if role == "" {
		role = entities.TokenRoleUser
	}

	// Create token entity
	now := time.Now()
	token := &entities.Token{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Name:       name,
		Key:        key,
		Status:     status,
		Role:       role,
		UsageCount: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.memoryRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"token_id": token.ID, "name": token.Name, "role": role}).Info("Token created")
	return token, nil
}

// GetTokenByID retrieves a token by ID
func (s *TokenService) GetTokenByID(ctx context.Context, id string) (*entities.Token, error) {
	return s.memoryRepo.GetByID(ctx, id)
}

// GetTokenByKey retrieves a token by its key
func (s *TokenService) GetTokenByKey(ctx context.Context, key string) (*entities.Token, error) {
	return s.memoryRepo.GetByKey(ctx, key)
}

// ListTokens retrieves tokens with optional filtering and pagination
// Pagination metadata is injected into the paging pointer
func (s *TokenService) ListTokens(
	ctx context.Context,
	query *dto.TokenQueryParams,
	paging *core.Paging,
) ([]*entities.Token, error) {
	// Get all tokens from memory
	allTokens, err := s.memoryRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Filter tokens based on query params
	filtered := make([]*entities.Token, 0)
	for _, token := range allTokens {
		// Filter by role
		if query.Role != "" && string(token.Role) != query.Role {
			continue
		}

		// Filter by status
		if query.Status != "" && string(token.Status) != query.Status {
			continue
		}

		// Search by name or key (case-insensitive)
		if query.Search != "" {
			searchLower := strings.ToLower(query.Search)
			nameLower := strings.ToLower(token.Name)
			keyLower := strings.ToLower(token.Key)
			if !strings.Contains(nameLower, searchLower) && !strings.Contains(keyLower, searchLower) {
				continue
			}
		}

		filtered = append(filtered, token)
	}

	// Set total count
	paging.Total = int64(len(filtered))

	// Apply pagination
	offset := (paging.Page - 1) * paging.Limit
	limit := paging.Limit

	// Calculate pagination bounds
	start := offset
	end := offset + limit
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	// Return paginated slice
	return filtered[start:end], nil
}

// UpdateToken updates an existing token
func (s *TokenService) UpdateToken(
	ctx context.Context,
	id, name, key string,
	status entities.TokenStatus,
	role entities.TokenRole,
) (*entities.Token, error) {
	// Get existing token
	token, err := s.memoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}

	// Check if key is being changed and if it already exists in another token
	if token.Key != key {
		existingToken, err := s.memoryRepo.GetByKey(ctx, key)
		if err == nil && existingToken != nil && existingToken.ID != id {
			return nil, fmt.Errorf("token with key already exists")
		}
	}

	// Update fields
	token.Name = name
	token.Key = key
	token.Status = status
	token.Role = role
	token.UpdatedAt = time.Now()

	if err := s.memoryRepo.Update(ctx, token); err != nil {
		return nil, err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"token_id": token.ID}).Info("Token updated")
	return token, nil
}

// DeleteToken deletes a token by ID
func (s *TokenService) DeleteToken(ctx context.Context, id string) error {
	if err := s.memoryRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.markDirty()
	s.logger.Withs(sctx.Fields{"token_id": id}).Info("Token deleted")
	return nil
}

// ValidateToken validates a token key and returns the token if valid
func (s *TokenService) ValidateToken(ctx context.Context, key string) (*entities.Token, error) {
	token, err := s.memoryRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("token not found")
	}

	// Check if token is active
	if token.Status != entities.TokenStatusActive {
		return nil, fmt.Errorf("token is not active")
	}

	// Update usage count and last used time
	token.IncrementUsage()
	if err := s.memoryRepo.Update(ctx, token); err != nil {
		s.logger.Withs(sctx.Fields{
			"token_id": token.ID,
			"error":    err,
		}).Warn("Failed to update token usage")
	} else {
		s.markDirty()
	}

	return token, nil
}
