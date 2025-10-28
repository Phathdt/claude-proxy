package entities

import (
	"fmt"
	"time"
)

// Session represents an active API session
type Session struct {
	ID          string    // Unique session identifier (UUID)
	AccountID   string    // Associated account ID
	TokenID     string    // API token used for this session
	UserAgent   string    // User agent string
	IPAddress   string    // Client IP address
	CreatedAt   time.Time // When session was created
	LastSeenAt  time.Time // Last activity timestamp
	ExpiresAt   time.Time // When session expires
	IsActive    bool      // Whether session is currently active
	RequestPath string    // Last request path (for debugging)
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UpdateLastSeen updates the last seen timestamp
func (s *Session) UpdateLastSeen() {
	s.LastSeenAt = time.Now()
}

// Refresh extends the session expiration
func (s *Session) Refresh(ttl time.Duration) {
	s.ExpiresAt = time.Now().Add(ttl)
	s.LastSeenAt = time.Now()
}

// Deactivate marks the session as inactive
func (s *Session) Deactivate() {
	s.IsActive = false
}

// ToMap converts session to map for Redis storage
func (s *Session) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":           s.ID,
		"account_id":   s.AccountID,
		"token_id":     s.TokenID,
		"user_agent":   s.UserAgent,
		"ip_address":   s.IPAddress,
		"created_at":   s.CreatedAt.Unix(),
		"last_seen_at": s.LastSeenAt.Unix(),
		"expires_at":   s.ExpiresAt.Unix(),
		"is_active":    s.IsActive,
		"request_path": s.RequestPath,
	}
}

// SessionFromMap creates a session from Redis map
func SessionFromMap(data map[string]string) *Session {
	return &Session{
		ID:          data["id"],
		AccountID:   data["account_id"],
		TokenID:     data["token_id"],
		UserAgent:   data["user_agent"],
		IPAddress:   data["ip_address"],
		CreatedAt:   parseUnixTime(data["created_at"]),
		LastSeenAt:  parseUnixTime(data["last_seen_at"]),
		ExpiresAt:   parseUnixTime(data["expires_at"]),
		IsActive:    data["is_active"] == "true",
		RequestPath: data["request_path"],
	}
}

func parseUnixTime(s string) time.Time {
	var timestamp int64
	fmt.Sscanf(s, "%d", &timestamp)
	return time.Unix(timestamp, 0)
}
