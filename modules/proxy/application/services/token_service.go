package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"
)

// TokenService implements the token management business logic
type TokenService struct {
	tokenRepo interfaces.TokenRepository
}

// NewTokenService creates a new token service
func NewTokenService(tokenRepo interfaces.TokenRepository) interfaces.TokenService {
	return &TokenService{
		tokenRepo: tokenRepo,
	}
}

// CreateToken creates a new API token
func (s *TokenService) CreateToken(ctx context.Context, name, key string, status entities.TokenStatus) (*entities.Token, error) {
	if name == "" {
		return nil, fmt.Errorf("token name is required")
	}
	if key == "" {
		return nil, fmt.Errorf("token key is required")
	}

	token := &entities.Token{
		Name:       name,
		Key:        key,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UsageCount: 0,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

// GetToken retrieves a token by ID
func (s *TokenService) GetToken(ctx context.Context, id string) (*entities.Token, error) {
	return s.tokenRepo.GetByID(ctx, id)
}

// ListTokens retrieves all tokens
func (s *TokenService) ListTokens(ctx context.Context) ([]*entities.Token, error) {
	return s.tokenRepo.List(ctx)
}

// UpdateToken updates an existing token
func (s *TokenService) UpdateToken(ctx context.Context, id, name, key string, status entities.TokenStatus) (*entities.Token, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	token.Update(name, key, status)

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}

	return token, nil
}

// DeleteToken deletes a token
func (s *TokenService) DeleteToken(ctx context.Context, id string) error {
	return s.tokenRepo.Delete(ctx, id)
}

// ValidateToken validates a token and increments usage
func (s *TokenService) ValidateToken(ctx context.Context, key string) (*entities.Token, error) {
	token, err := s.tokenRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	if !token.IsActive() {
		return nil, fmt.Errorf("token is inactive")
	}

	// Increment usage
	token.IncrementUsage()

	// Save updated token
	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to update token usage: %w", err)
	}

	return token, nil
}

// GenerateKey generates a random API key
func (s *TokenService) GenerateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "tk_" + hex.EncodeToString(bytes), nil
}
