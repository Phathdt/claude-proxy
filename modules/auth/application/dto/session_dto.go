package dto

import (
	"time"

	"claude-proxy/modules/auth/domain/entities"
)

// ============================================================================
// Persistence DTOs (for JSON file storage)
// ============================================================================

// SessionPersistenceDTO represents the JSON structure for session persistence
type SessionPersistenceDTO struct {
	ID          string `json:"id"`
	TokenID     string `json:"token_id"`
	UserAgent   string `json:"user_agent"`
	IPAddress   string `json:"ip_address"`
	CreatedAt   string `json:"created_at"`   // RFC3339/ISO 8601 datetime
	LastSeenAt  string `json:"last_seen_at"` // RFC3339/ISO 8601 datetime
	ExpiresAt   string `json:"expires_at"`   // RFC3339/ISO 8601 datetime
	IsActive    bool   `json:"is_active"`
	RequestPath string `json:"request_path,omitempty"`
}

// ToSessionPersistenceDTO converts session entity to persistence DTO
func ToSessionPersistenceDTO(session *entities.Session) *SessionPersistenceDTO {
	return &SessionPersistenceDTO{
		ID:          session.ID,
		TokenID:     session.TokenID,
		UserAgent:   session.UserAgent,
		IPAddress:   session.IPAddress,
		CreatedAt:   session.CreatedAt.Format(time.RFC3339),
		LastSeenAt:  session.LastSeenAt.Format(time.RFC3339),
		ExpiresAt:   session.ExpiresAt.Format(time.RFC3339),
		IsActive:    session.IsActive,
		RequestPath: session.RequestPath,
	}
}

// FromSessionPersistenceDTO converts persistence DTO to session entity
func FromSessionPersistenceDTO(dto *SessionPersistenceDTO) *entities.Session {
	createdAt, _ := time.Parse(time.RFC3339, dto.CreatedAt)
	lastSeenAt, _ := time.Parse(time.RFC3339, dto.LastSeenAt)
	expiresAt, _ := time.Parse(time.RFC3339, dto.ExpiresAt)

	return &entities.Session{
		ID:          dto.ID,
		TokenID:     dto.TokenID,
		UserAgent:   dto.UserAgent,
		IPAddress:   dto.IPAddress,
		CreatedAt:   createdAt,
		LastSeenAt:  lastSeenAt,
		ExpiresAt:   expiresAt,
		IsActive:    dto.IsActive,
		RequestPath: dto.RequestPath,
	}
}

// ============================================================================
// API Response DTOs (for HTTP responses)
// ============================================================================

// SessionResponse represents a session in API responses
type SessionResponse struct {
	ID          string `json:"id"`
	TokenID     string `json:"token_id"`
	UserAgent   string `json:"user_agent"`
	IPAddress   string `json:"ip_address"`
	CreatedAt   string `json:"created_at"`   // RFC3339/ISO 8601 datetime
	LastSeenAt  string `json:"last_seen_at"` // RFC3339/ISO 8601 datetime
	ExpiresAt   string `json:"expires_at"`   // RFC3339/ISO 8601 datetime
	IsActive    bool   `json:"is_active"`
	RequestPath string `json:"request_path"`
}

// ListSessionsResponse represents a list of sessions
type ListSessionsResponse struct {
	Sessions []*SessionResponse `json:"sessions"`
	Total    int                `json:"total"`
}

// RevokeSessionRequest represents a request to revoke a session
type RevokeSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// RevokeSessionResponse represents a response to session revocation
type RevokeSessionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
