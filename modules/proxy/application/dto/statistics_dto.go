package dto

// SystemHealth represents the overall health status of the system
type SystemHealth string

const (
	SystemHealthHealthy   SystemHealth = "healthy"   // ≥2 active accounts
	SystemHealthDegraded  SystemHealth = "degraded"  // 1 active account
	SystemHealthUnhealthy SystemHealth = "unhealthy" // 0 active accounts
)

// StatisticsResponse represents the system statistics response
type StatisticsResponse struct {
	// Account counts by status
	TotalAccounts        int `json:"total_accounts"`
	ActiveAccounts       int `json:"active_accounts"`
	InactiveAccounts     int `json:"inactive_accounts"`
	RateLimitedAccounts  int `json:"rate_limited_accounts"`
	InvalidAccounts      int `json:"invalid_accounts"`

	// Token health metrics
	AccountsNeedingRefresh int     `json:"accounts_needing_refresh"`
	OldestTokenAgeHours    float64 `json:"oldest_token_age_hours"`

	// System health
	SystemHealth SystemHealth `json:"system_health"`
}
