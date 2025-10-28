package dto

// SessionResponse represents a session in API responses
type SessionResponse struct {
	ID          string `json:"id"`
	TokenID     string `json:"token_id"`
	UserAgent   string `json:"user_agent"`
	IPAddress   string `json:"ip_address"`
	CreatedAt   int64  `json:"created_at"`
	LastSeenAt  int64  `json:"last_seen_at"`
	ExpiresAt   int64  `json:"expires_at"`
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
