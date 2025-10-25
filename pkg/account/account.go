package account

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Account represents a Claude account with OAuth tokens
type Account struct {
	OrganizationUUID string    `json:"organization_uuid"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        int64     `json:"expires_at"` // Unix timestamp
	Status           string    `json:"status"`     // "valid" or "invalid"
	CreatedAt        int64     `json:"created_at"`
	UpdatedAt        int64     `json:"updated_at"`
}

// TokenRefresher defines the interface for refreshing tokens
type TokenRefresher interface {
	RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, expiresIn int, err error)
}

// Manager handles account persistence and token refresh
type Manager struct {
	dataFolder     string
	refresher      TokenRefresher
	mu             sync.Mutex
	account        *Account
	refreshing     bool
}

// NewManager creates a new account manager
func NewManager(dataFolder string, refresher TokenRefresher) *Manager {
	return &Manager{
		dataFolder: expandPath(dataFolder),
		refresher:  refresher,
	}
}

// Initialize initializes the account manager by creating data folder and loading account
func (m *Manager) Initialize() error {
	// Create data folder if not exists
	if err := os.MkdirAll(m.dataFolder, 0700); err != nil {
		return fmt.Errorf("failed to create data folder: %w", err)
	}

	// Try to load existing account
	if err := m.loadAccount(); err != nil {
		// Account file doesn't exist yet, that's ok
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load account: %w", err)
		}
	}

	return nil
}

// SaveAccount saves a new account
func (m *Manager) SaveAccount(account *Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().Unix()
	if account.CreatedAt == 0 {
		account.CreatedAt = now
	}
	account.UpdatedAt = now
	account.Status = "valid"

	if err := m.writeAccount(account); err != nil {
		return err
	}

	m.account = account
	return nil
}

// GetValidToken returns a valid access token, refreshing if necessary
func (m *Manager) GetValidToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.account == nil {
		return "", fmt.Errorf("no account configured")
	}

	// Check if token is expired or will expire in the next 60 seconds
	if time.Now().Unix() >= m.account.ExpiresAt-60 {
		// Refresh token
		if m.refreshing {
			return "", fmt.Errorf("token refresh already in progress")
		}

		m.refreshing = true
		defer func() { m.refreshing = false }()

		if err := m.refreshTokenLocked(ctx); err != nil {
			m.account.Status = "invalid"
			_ = m.writeAccount(m.account)
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	return m.account.AccessToken, nil
}

// GetAccount returns the current account
func (m *Manager) GetAccount() *Account {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.account == nil {
		return nil
	}

	// Return a copy
	account := *m.account
	return &account
}

// GetStatus returns account status
func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.account == nil {
		return map[string]interface{}{
			"account_valid": false,
		}
	}

	return map[string]interface{}{
		"account_valid": m.account.Status == "valid",
		"expires_at":    m.account.ExpiresAt,
		"organization":  m.account.OrganizationUUID,
	}
}

// refreshTokenLocked refreshes the access token (must be called with lock held)
func (m *Manager) refreshTokenLocked(ctx context.Context) error {
	accessToken, refreshToken, expiresIn, err := m.refresher.RefreshAccessToken(ctx, m.account.RefreshToken)
	if err != nil {
		return err
	}

	// Update account with new tokens
	m.account.AccessToken = accessToken
	m.account.RefreshToken = refreshToken
	m.account.ExpiresAt = time.Now().Unix() + int64(expiresIn)
	m.account.UpdatedAt = time.Now().Unix()
	m.account.Status = "valid"

	// Save to disk
	return m.writeAccount(m.account)
}

// loadAccount loads account from disk
func (m *Manager) loadAccount() error {
	accountPath := filepath.Join(m.dataFolder, "account.json")
	data, err := os.ReadFile(accountPath)
	if err != nil {
		return err
	}

	var account Account
	if err := json.Unmarshal(data, &account); err != nil {
		return fmt.Errorf("failed to unmarshal account: %w", err)
	}

	m.account = &account
	return nil
}

// writeAccount writes account to disk atomically
func (m *Manager) writeAccount(account *Account) error {
	accountPath := filepath.Join(m.dataFolder, "account.json")

	// Marshal account to JSON
	data, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	// Write to temporary file first
	tmpPath := accountPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write account file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, accountPath); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename account file: %w", err)
	}

	return nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
