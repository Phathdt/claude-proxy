package account

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AppAccount represents a single Claude app account
type AppAccount struct {
	ID               string `json:"id"`                // Unique identifier
	Name             string `json:"name"`              // Friendly name
	OrganizationUUID string `json:"organization_uuid"` // Claude org UUID
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresAt        int64  `json:"expires_at"` // Unix timestamp
	Status           string `json:"status"`     // "active" or "inactive"
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

// MultiAccountManager manages multiple Claude accounts
type MultiAccountManager struct {
	dataFolder string
	refresher  TokenRefresher
	mu         sync.RWMutex
	accounts   map[string]*AppAccount // id -> account
}

// NewMultiAccountManager creates a new multi-account manager
func NewMultiAccountManager(dataFolder string, refresher TokenRefresher) *MultiAccountManager {
	return &MultiAccountManager{
		dataFolder: expandPath(dataFolder),
		refresher:  refresher,
		accounts:   make(map[string]*AppAccount),
	}
}

// Initialize loads all accounts from disk
func (m *MultiAccountManager) Initialize() error {
	if err := os.MkdirAll(m.dataFolder, 0700); err != nil {
		return fmt.Errorf("failed to create data folder: %w", err)
	}

	return m.loadAccounts()
}

// CreateAccount creates a new account
func (m *MultiAccountManager) CreateAccount(name, orgUUID, accessToken, refreshToken string, expiresIn int) (*AppAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique ID
	id := generateID()
	now := time.Now().Unix()

	account := &AppAccount{
		ID:               id,
		Name:             name,
		OrganizationUUID: orgUUID,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        now + int64(expiresIn),
		Status:           "active",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	m.accounts[id] = account

	if err := m.saveAccounts(); err != nil {
		delete(m.accounts, id)
		return nil, err
	}

	return account, nil
}

// GetAccount retrieves an account by ID
func (m *MultiAccountManager) GetAccount(id string) (*AppAccount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	account, exists := m.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found: %s", id)
	}

	// Return a copy
	accountCopy := *account
	return &accountCopy, nil
}

// ListAccounts returns all accounts
func (m *MultiAccountManager) ListAccounts() []*AppAccount {
	m.mu.RLock()
	defer m.mu.RUnlock()

	accounts := make([]*AppAccount, 0, len(m.accounts))
	for _, account := range m.accounts {
		accountCopy := *account
		accounts = append(accounts, &accountCopy)
	}

	return accounts
}

// UpdateAccount updates an account
func (m *MultiAccountManager) UpdateAccount(id, name, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	account, exists := m.accounts[id]
	if !exists {
		return fmt.Errorf("account not found: %s", id)
	}

	if name != "" {
		account.Name = name
	}
	if status != "" {
		account.Status = status
	}
	account.UpdatedAt = time.Now().Unix()

	return m.saveAccounts()
}

// DeleteAccount deletes an account
func (m *MultiAccountManager) DeleteAccount(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.accounts[id]; !exists {
		return fmt.Errorf("account not found: %s", id)
	}

	delete(m.accounts, id)
	return m.saveAccounts()
}

// GetValidToken returns a valid access token for an account, refreshing if needed
func (m *MultiAccountManager) GetValidToken(ctx context.Context, id string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	account, exists := m.accounts[id]
	if !exists {
		return "", fmt.Errorf("account not found: %s", id)
	}

	// Check if token needs refresh (60s buffer)
	if time.Now().Unix() >= account.ExpiresAt-60 {
		if err := m.refreshTokenLocked(ctx, account); err != nil {
			account.Status = "inactive"
			_ = m.saveAccounts()
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	return account.AccessToken, nil
}

// refreshTokenLocked refreshes an account's token (must be called with lock held)
func (m *MultiAccountManager) refreshTokenLocked(ctx context.Context, account *AppAccount) error {
	accessToken, refreshToken, expiresIn, err := m.refresher.RefreshAccessToken(ctx, account.RefreshToken)
	if err != nil {
		return err
	}

	account.AccessToken = accessToken
	account.RefreshToken = refreshToken
	account.ExpiresAt = time.Now().Unix() + int64(expiresIn)
	account.UpdatedAt = time.Now().Unix()
	account.Status = "active"

	return m.saveAccounts()
}

// loadAccounts loads all accounts from disk
func (m *MultiAccountManager) loadAccounts() error {
	accountsPath := filepath.Join(m.dataFolder, "accounts.json")
	data, err := os.ReadFile(accountsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No accounts yet
		}
		return fmt.Errorf("failed to read accounts file: %w", err)
	}

	var accounts []*AppAccount
	if err := json.Unmarshal(data, &accounts); err != nil {
		return fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	m.accounts = make(map[string]*AppAccount)
	for _, account := range accounts {
		m.accounts[account.ID] = account
	}

	return nil
}

// saveAccounts saves all accounts to disk atomically
func (m *MultiAccountManager) saveAccounts() error {
	accountsPath := filepath.Join(m.dataFolder, "accounts.json")

	// Convert map to slice
	accounts := make([]*AppAccount, 0, len(m.accounts))
	for _, account := range m.accounts {
		accounts = append(accounts, account)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	// Write to temporary file first
	tmpPath := accountsPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write accounts file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, accountsPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename accounts file: %w", err)
	}

	return nil
}

// generateID generates a unique ID for an account
func generateID() string {
	return fmt.Sprintf("app_%d", time.Now().UnixNano())
}
