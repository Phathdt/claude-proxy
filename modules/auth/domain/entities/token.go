package entities

import "time"

// Token represents an API token for authentication
type Token struct {
	ID         string
	Name       string
	Key        string
	Status     TokenStatus
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
)

// IsActive returns true if the token is active
func (t *Token) IsActive() bool {
	return t.Status == TokenStatusActive
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

// Update updates the token's name and key
func (t *Token) Update(name, key string, status TokenStatus) {
	t.Name = name
	t.Key = key
	t.Status = status
	t.UpdatedAt = time.Now()
}
