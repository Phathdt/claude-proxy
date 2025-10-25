package dto

import "claude-proxy/modules/proxy/domain/entities"

// CreateAccountRequest represents the request to create an account
type CreateAccountRequest struct {
	Name  string `json:"name" binding:"required"`
	OrgID string `json:"org_id,omitempty"`
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
}

// AccountResponse represents the account response
type AccountResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	OrganizationUUID string `json:"organization_uuid"`
	ExpiresAt        int64  `json:"expires_at"`
	Status           string `json:"status"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

// ToAccountResponse converts entity to response DTO (without sensitive tokens)
func ToAccountResponse(account *entities.Account) *AccountResponse {
	return &AccountResponse{
		ID:               account.ID,
		Name:             account.Name,
		OrganizationUUID: account.OrganizationUUID,
		ExpiresAt:        account.ExpiresAt.Unix(),
		Status:           string(account.Status),
		CreatedAt:        account.CreatedAt.Unix(),
		UpdatedAt:        account.UpdatedAt.Unix(),
	}
}

// ToAccountResponses converts entity slice to response DTO slice
func ToAccountResponses(accounts []*entities.Account) []*AccountResponse {
	responses := make([]*AccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = ToAccountResponse(account)
	}
	return responses
}
