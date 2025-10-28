package dto

import (
	"claude-proxy/modules/auth/domain/entities"
)

// TokenQueryParams represents query parameters for listing tokens
type TokenQueryParams struct {
	Role   string `form:"role"`   // Filter by role: "admin", "user", or empty for all
	Status string `form:"status"` // Filter by status: "active", "inactive", or empty for all
	Search string `form:"search"` // Search by name or key
}

// CreateTokenRequest represents the request to create a token
type CreateTokenRequest struct {
	Name   string `json:"name"   binding:"required"`
	Key    string `json:"key"    binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
	Role   string `json:"role"   binding:"omitempty,oneof=user admin"` // Default: user
}

// UpdateTokenRequest represents the request to update a token
type UpdateTokenRequest struct {
	Name   string `json:"name"   binding:"required"`
	Key    string `json:"key"    binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
	Role   string `json:"role"   binding:"required,oneof=user admin"`
}

// TokenResponse represents the token response
type TokenResponse struct {
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

// ToTokenResponse converts entity to response DTO
func ToTokenResponse(token *entities.Token) *TokenResponse {
	resp := &TokenResponse{
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
