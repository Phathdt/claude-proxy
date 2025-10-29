package entities

import "time"

// Token represents an API token for authentication
type Token struct {
	ID         string
	Name       string
	Key        string
	Status     TokenStatus
	Role       TokenRole
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UsageCount int
	LastUsedAt *time.Time
}

// TokenStatus represents the status of a token
type TokenStatus string

const (
	TokenStatusActive   TokenStatus = "active"
	TokenStatusInactive TokenStatus = "inactive"
	TokenStatusRevoked  TokenStatus = "revoked"
)

// TokenRole represents the role of a token
type TokenRole string

const (
	TokenRoleUser  TokenRole = "user"  // Regular API access
	TokenRoleAdmin TokenRole = "admin" // Admin UI access
)

// IsActive returns true if the token is active
func (t *Token) IsActive() bool {
	return t.Status == TokenStatusActive
}

// IsAdmin returns true if the token has admin role
func (t *Token) IsAdmin() bool {
	return t.Role == TokenRoleAdmin
}

// IncrementUsage increments the usage count and updates last used time
func (t *Token) IncrementUsage() {
	t.UsageCount++
	now := time.Now()
	t.LastUsedAt = &now
}

// Deactivate deactivates the token
func (t *Token) Deactivate() {
	t.Status = TokenStatusInactive
}

// Activate activates the token
func (t *Token) Activate() {
	t.Status = TokenStatusActive
}

// Revoke revokes the token (permanent deactivation)
func (t *Token) Revoke() {
	t.Status = TokenStatusRevoked
}

// Update updates the token's name, key, status and role
func (t *Token) Update(name, key string, status TokenStatus, role TokenRole) {
	t.Name = name
	t.Key = key
	t.Status = status
	t.Role = role
	t.UpdatedAt = time.Now()
}
