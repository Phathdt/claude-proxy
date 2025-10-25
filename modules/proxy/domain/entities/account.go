package entities

import "time"

// Account represents a Claude OAuth account
type Account struct {
	ID               string
	Name             string
	OrganizationUUID string
	AccessToken      string
	RefreshToken     string
	ExpiresAt        time.Time
	Status           AccountStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AccountStatus represents the status of an app account
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
)

// IsActive returns true if the account is active
func (a *Account) IsActive() bool {
	return a.Status == AccountStatusActive
}

// IsExpired returns true if the access token is expired
func (a *Account) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// NeedsRefresh returns true if the token needs refresh (60s buffer)
func (a *Account) NeedsRefresh() bool {
	return time.Now().After(a.ExpiresAt.Add(-60 * time.Second))
}

// UpdateTokens updates the access token, refresh token and expiry
func (a *Account) UpdateTokens(accessToken, refreshToken string, expiresIn int) {
	a.AccessToken = accessToken
	a.RefreshToken = refreshToken
	a.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	a.UpdatedAt = time.Now()
	a.Status = AccountStatusActive
}

// Deactivate marks the account as inactive
func (a *Account) Deactivate() {
	a.Status = AccountStatusInactive
	a.UpdatedAt = time.Now()
}

// Activate marks the account as active
func (a *Account) Activate() {
	a.Status = AccountStatusActive
	a.UpdatedAt = time.Now()
}

// Update updates the account's name and status
func (a *Account) Update(name string, status AccountStatus) {
	if name != "" {
		a.Name = name
	}
	if status != "" {
		a.Status = status
	}
	a.UpdatedAt = time.Now()
}
