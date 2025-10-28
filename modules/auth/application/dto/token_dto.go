package dto

import "claude-proxy/modules/auth/domain/entities"

// CreateTokenRequest represents the request to create a token
type CreateTokenRequest struct {
	Name   string `json:"name"   binding:"required"`
	Key    string `json:"key"    binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// UpdateTokenRequest represents the request to update a token
type UpdateTokenRequest struct {
	Name   string `json:"name"   binding:"required"`
	Key    string `json:"key"    binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	Status     string `json:"status"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	UsageCount int    `json:"usage_count"`
	LastUsedAt *int64 `json:"last_used_at,omitempty"`
}

// ToTokenResponse converts entity to response DTO
func ToTokenResponse(token *entities.Token) *TokenResponse {
	resp := &TokenResponse{
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
