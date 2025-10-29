package dto

import (
	"time"

	"claude-proxy/modules/auth/domain/entities"
)

// RFC3339 is the datetime format for API responses and persistence (ISO 8601)
const RFC3339 = time.RFC3339

// ============================================================================
// Persistence DTOs (for JSON file storage)
// ============================================================================

// AccountPersistenceDTO represents the JSON structure for account persistence
type AccountPersistenceDTO struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	OrganizationUUID string  `json:"organization_uuid"`
	AccessToken      string  `json:"access_token"`
	RefreshToken     string  `json:"refresh_token"`
	ExpiresAt        string  `json:"expires_at"` // RFC3339/ISO 8601 datetime
	Status           string  `json:"status"`
	RateLimitedUntil *string `json:"rate_limited_until,omitempty"` // RFC3339/ISO 8601 datetime
	LastRefreshError string  `json:"last_refresh_error,omitempty"`
	CreatedAt        string  `json:"created_at"` // RFC3339/ISO 8601 datetime
	UpdatedAt        string  `json:"updated_at"` // RFC3339/ISO 8601 datetime
}

// ToAccountPersistenceDTO converts entity to persistence DTO (includes sensitive tokens)
func ToAccountPersistenceDTO(account *entities.Account) *AccountPersistenceDTO {
	dto := &AccountPersistenceDTO{
		ID:               account.ID,
		Name:             account.Name,
		OrganizationUUID: account.OrganizationUUID,
		AccessToken:      account.AccessToken,
		RefreshToken:     account.RefreshToken,
		ExpiresAt:        account.ExpiresAt.Format(RFC3339),
		Status:           string(account.Status),
		LastRefreshError: account.LastRefreshError,
		CreatedAt:        account.CreatedAt.Format(RFC3339),
		UpdatedAt:        account.UpdatedAt.Format(RFC3339),
	}

	// Convert RateLimitedUntil pointer
	if account.RateLimitedUntil != nil {
		timestamp := account.RateLimitedUntil.Format(RFC3339)
		dto.RateLimitedUntil = &timestamp
	}

	return dto
}

// FromAccountPersistenceDTO converts persistence DTO to entity
func FromAccountPersistenceDTO(dto *AccountPersistenceDTO) *entities.Account {
	expiresAt, _ := time.Parse(RFC3339, dto.ExpiresAt)
	createdAt, _ := time.Parse(RFC3339, dto.CreatedAt)
	updatedAt, _ := time.Parse(RFC3339, dto.UpdatedAt)

	account := &entities.Account{
		ID:               dto.ID,
		Name:             dto.Name,
		OrganizationUUID: dto.OrganizationUUID,
		AccessToken:      dto.AccessToken,
		RefreshToken:     dto.RefreshToken,
		ExpiresAt:        expiresAt,
		Status:           entities.AccountStatus(dto.Status),
		LastRefreshError: dto.LastRefreshError,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}

	// Convert RateLimitedUntil pointer
	if dto.RateLimitedUntil != nil {
		t, _ := time.Parse(RFC3339, *dto.RateLimitedUntil)
		account.RateLimitedUntil = &t
	}

	return account
}

// ============================================================================
// API Response DTOs (for HTTP responses - no sensitive data)
// ============================================================================

// CreateAccountRequest represents the request to create an account
type CreateAccountRequest struct {
	Name  string `json:"name"             binding:"required"`
	OrgID string `json:"org_id,omitempty"`
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive rate_limited invalid"`
}

// AccountResponse represents the account response
type AccountResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	OrganizationUUID string  `json:"organization_uuid"`
	ExpiresAt        string  `json:"expires_at"` // RFC3339/ISO 8601 datetime
	Status           string  `json:"status"`
	RateLimitedUntil *string `json:"rate_limited_until,omitempty"` // RFC3339/ISO 8601 datetime, nil if not rate limited
	LastRefreshError string  `json:"last_refresh_error,omitempty"` // Error message from last refresh attempt
	CreatedAt        string  `json:"created_at"`                   // RFC3339/ISO 8601 datetime
	UpdatedAt        string  `json:"updated_at"`                   // RFC3339/ISO 8601 datetime
}

// ToAccountResponse converts entity to response DTO (without sensitive tokens)
func ToAccountResponse(account *entities.Account) *AccountResponse {
	resp := &AccountResponse{
		ID:               account.ID,
		Name:             account.Name,
		OrganizationUUID: account.OrganizationUUID,
		ExpiresAt:        account.ExpiresAt.Format(RFC3339),
		Status:           string(account.Status),
		LastRefreshError: account.LastRefreshError,
		CreatedAt:        account.CreatedAt.Format(RFC3339),
		UpdatedAt:        account.UpdatedAt.Format(RFC3339),
	}

	// Include rate limited until if present
	if account.RateLimitedUntil != nil {
		timestamp := account.RateLimitedUntil.Format(RFC3339)
		resp.RateLimitedUntil = &timestamp
	}

	return resp
}

// ToAccountResponses converts entity slice to response DTO slice
func ToAccountResponses(accounts []*entities.Account) []*AccountResponse {
	responses := make([]*AccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = ToAccountResponse(account)
	}
	return responses
}
