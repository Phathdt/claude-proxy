package dto

import (
	"time"

	"claude-proxy/modules/auth/domain/entities"
)

// ============================================================================
// Persistence DTOs (for JSON file storage)
// ============================================================================

// TokenPersistenceDTO represents the JSON structure for token persistence
type TokenPersistenceDTO struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Key        string  `json:"key"`
	Status     string  `json:"status"`
	Role       string  `json:"role"`       // user or admin
	CreatedAt  string  `json:"created_at"` // RFC3339/ISO 8601 datetime
	UpdatedAt  string  `json:"updated_at"` // RFC3339/ISO 8601 datetime
	UsageCount int     `json:"usage_count"`
	LastUsedAt *string `json:"last_used_at,omitempty"` // RFC3339/ISO 8601 datetime
}

// ToTokenPersistenceDTO converts token entity to persistence DTO (includes sensitive data)
func ToTokenPersistenceDTO(token *entities.Token) *TokenPersistenceDTO {
	dto := &TokenPersistenceDTO{
		ID:         token.ID,
		Name:       token.Name,
		Key:        token.Key,
		Status:     string(token.Status),
		Role:       string(token.Role),
		CreatedAt:  token.CreatedAt.Format(RFC3339),
		UpdatedAt:  token.UpdatedAt.Format(RFC3339),
		UsageCount: token.UsageCount,
	}

	if token.LastUsedAt != nil {
		lastUsed := token.LastUsedAt.Format(RFC3339)
		dto.LastUsedAt = &lastUsed
	}

	return dto
}

// FromTokenPersistenceDTO converts persistence DTO to token entity
func FromTokenPersistenceDTO(dto *TokenPersistenceDTO) *entities.Token {
	createdAt, _ := time.Parse(RFC3339, dto.CreatedAt)
	updatedAt, _ := time.Parse(RFC3339, dto.UpdatedAt)

	// Default role to user if not set (backward compatibility)
	role := entities.TokenRole(dto.Role)
	if role == "" {
		role = entities.TokenRoleUser
	}

	token := &entities.Token{
		ID:         dto.ID,
		Name:       dto.Name,
		Key:        dto.Key,
		Status:     entities.TokenStatus(dto.Status),
		Role:       role,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		UsageCount: dto.UsageCount,
	}

	if dto.LastUsedAt != nil {
		lastUsed, _ := time.Parse(RFC3339, *dto.LastUsedAt)
		token.LastUsedAt = &lastUsed
	}

	return token
}

// ============================================================================
// Query DTOs (for filtering and pagination)
// ============================================================================

// TokenQueryParams represents query parameters for listing tokens
type TokenQueryParams struct {
	Role   string `form:"role"`   // Filter by role (user/admin)
	Status string `form:"status"` // Filter by status (active/inactive/revoked)
	Search string `form:"search"` // Search by name or key
}

// ============================================================================
// API Request DTOs (for HTTP requests)
// ============================================================================

// CreateTokenRequest represents the request to create a token
type CreateTokenRequest struct {
	Name   string `json:"name"   binding:"required"`
	Key    string `json:"key"    binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive revoked"`
	Role   string `json:"role"   binding:"required,oneof=user admin"`
}

// UpdateTokenRequest represents the request to update a token
type UpdateTokenRequest struct {
	Name   *string `json:"name,omitempty"`
	Key    *string `json:"key,omitempty"`
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive revoked"`
	Role   *string `json:"role,omitempty"   binding:"omitempty,oneof=user admin"`
}

// ============================================================================
// API Response DTOs (for HTTP responses - no sensitive data)
// ============================================================================

// TokenResponse represents the token response
type TokenResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Key        string  `json:"key"` // Masked for security (first 6 + last 6 chars)
	Status     string  `json:"status"`
	Role       string  `json:"role"`
	CreatedAt  string  `json:"created_at"` // RFC3339/ISO 8601 datetime
	UpdatedAt  string  `json:"updated_at"` // RFC3339/ISO 8601 datetime
	UsageCount int     `json:"usage_count"`
	LastUsedAt *string `json:"last_used_at,omitempty"` // RFC3339/ISO 8601 datetime
}

// maskKey masks the API key showing only first 6 and last 6 characters
func maskKey(key string) string {
	if len(key) <= 12 {
		return "***"
	}
	return key[:6] + "***" + key[len(key)-6:]
}

// ToTokenResponse converts entity to response DTO with masked key
func ToTokenResponse(token *entities.Token) *TokenResponse {
	resp := &TokenResponse{
		ID:         token.ID,
		Name:       token.Name,
		Key:        maskKey(token.Key),
		Status:     string(token.Status),
		Role:       string(token.Role),
		CreatedAt:  token.CreatedAt.Format(RFC3339),
		UpdatedAt:  token.UpdatedAt.Format(RFC3339),
		UsageCount: token.UsageCount,
	}

	if token.LastUsedAt != nil {
		lastUsed := token.LastUsedAt.Format(RFC3339)
		resp.LastUsedAt = &lastUsed
	}

	return resp
}

// ToTokenResponseWithFullKey converts entity to response DTO with full key (use only for Create)
func ToTokenResponseWithFullKey(token *entities.Token) *TokenResponse {
	resp := &TokenResponse{
		ID:         token.ID,
		Name:       token.Name,
		Key:        token.Key, // Full key, not masked
		Status:     string(token.Status),
		Role:       string(token.Role),
		CreatedAt:  token.CreatedAt.Format(RFC3339),
		UpdatedAt:  token.UpdatedAt.Format(RFC3339),
		UsageCount: token.UsageCount,
	}

	if token.LastUsedAt != nil {
		lastUsed := token.LastUsedAt.Format(RFC3339)
		resp.LastUsedAt = &lastUsed
	}

	return resp
}

// ToTokenResponses converts entity slice to response DTO slice
func ToTokenResponses(tokens []*entities.Token) []*TokenResponse {
	responses := make([]*TokenResponse, len(tokens))
	for i, token := range tokens {
		responses[i] = ToTokenResponse(token)
	}
	return responses
}
