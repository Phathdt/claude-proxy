package entities

import "time"

// Account represents a Claude OAuth account
type Account struct {
	ID               string
	Name             string
	OrganizationUUID string
	AccessToken      string
	RefreshToken     string
	ExpiresAt        time.Time // When access token expires
	RefreshAt        time.Time // When tokens were last refreshed
	Status           AccountStatus
	RateLimitedUntil *time.Time // When rate limit expires (nil if not rate limited)
	LastRefreshError string     // Last error message from token refresh attempt
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AccountStatus represents the status of an app account
type AccountStatus string

const (
	AccountStatusActive      AccountStatus = "active"       // Healthy and available
	AccountStatusInactive    AccountStatus = "inactive"     // Manually disabled
	AccountStatusRateLimited AccountStatus = "rate_limited" // Temporarily rate limited
	AccountStatusInvalid     AccountStatus = "invalid"      // Auth revoked/invalid
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
// Also clears error state and rate limit, marking account as active
func (a *Account) UpdateTokens(accessToken, refreshToken string, expiresIn int) {
	a.AccessToken = accessToken
	a.RefreshToken = refreshToken
	a.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	a.RefreshAt = time.Now()
	a.UpdatedAt = time.Now()
	a.Status = AccountStatusActive
	a.RateLimitedUntil = nil // Clear rate limit
	a.LastRefreshError = ""  // Clear error on success
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

// UpdateRefreshError updates the account with a refresh error
func (a *Account) UpdateRefreshError(errMsg string) {
	a.LastRefreshError = errMsg
	a.UpdatedAt = time.Now()
}

// IsAvailableForProxy returns true if account can be used for proxying
func (a *Account) IsAvailableForProxy() bool {
	switch a.Status {
	case AccountStatusActive:
		return true
	case AccountStatusRateLimited:
		// Check if rate limit has expired
		if a.RateLimitedUntil != nil && time.Now().After(*a.RateLimitedUntil) {
			return true // Rate limit expired, can be recovered
		}
		return false
	case AccountStatusInvalid, AccountStatusInactive:
		return false
	default:
		return false
	}
}

// IsRateLimitExpired returns true if rate limit has expired
func (a *Account) IsRateLimitExpired() bool {
	if a.Status != AccountStatusRateLimited {
		return false
	}
	if a.RateLimitedUntil == nil {
		return true // No expiry set, consider expired
	}
	return time.Now().After(*a.RateLimitedUntil)
}

// MarkRateLimited marks account as rate limited until specified time
func (a *Account) MarkRateLimited(until time.Time, errMsg string) {
	a.Status = AccountStatusRateLimited
	a.RateLimitedUntil = &until
	a.LastRefreshError = errMsg
	a.UpdatedAt = time.Now()
}

// MarkInvalid marks account as invalid (auth revoked)
func (a *Account) MarkInvalid(errMsg string) {
	a.Status = AccountStatusInvalid
	a.RateLimitedUntil = nil
	a.LastRefreshError = errMsg
	a.UpdatedAt = time.Now()
}

// RecoverFromRateLimit marks account as active after rate limit expires
func (a *Account) RecoverFromRateLimit() {
	if a.Status == AccountStatusRateLimited && a.IsRateLimitExpired() {
		a.Status = AccountStatusActive
		a.RateLimitedUntil = nil
		a.LastRefreshError = ""
		a.UpdatedAt = time.Now()
	}
}
