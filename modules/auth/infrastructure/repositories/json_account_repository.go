package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"
)

// JSONAccountPersistenceRepository implements PersistenceRepository using JSON file storage
// This repository ONLY handles disk I/O, no in-memory caching
type JSONAccountPersistenceRepository struct {
	dataFolder string
	mu         sync.RWMutex // Only for file I/O concurrency control
}

// NewJSONAccountPersistenceRepository creates a new JSON persistence repository
func NewJSONAccountPersistenceRepository(dataFolder string) (interfaces.PersistenceRepository, error) {
	repo := &JSONAccountPersistenceRepository{
		dataFolder: expandPath(dataFolder),
	}

	// Create data folder if it doesn't exist
	if err := os.MkdirAll(repo.dataFolder, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create data folder: %w", err)
	}

	return repo, nil
}

// SaveAll persists all accounts to durable storage (batch operation)
func (r *JSONAccountPersistenceRepository) SaveAll(ctx context.Context, accounts []*entities.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	accountsFile := filepath.Join(r.dataFolder, "accounts.json")

	// Convert entities to DTOs
	dtos := make([]*dto.AccountPersistenceDTO, 0, len(accounts))
	for _, account := range accounts {
		dtos = append(dtos, dto.ToAccountPersistenceDTO(account))
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(dtos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	// Write to temporary file first (atomic write)
	tmpFile := accountsFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write accounts file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, accountsFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename accounts file: %w", err)
	}

	return nil
}

// LoadAll loads all accounts from durable storage
func (r *JSONAccountPersistenceRepository) LoadAll(ctx context.Context) ([]*entities.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accountsFile := filepath.Join(r.dataFolder, "accounts.json")

	data, err := os.ReadFile(accountsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entities.Account{}, nil // No accounts yet
		}
		return nil, fmt.Errorf("failed to read accounts file: %w", err)
	}

	// Try to parse as array first (new format)
	var dtos []*dto.AccountPersistenceDTO
	if err := json.Unmarshal(data, &dtos); err == nil && len(dtos) > 0 {
		// Successfully parsed as array
		accounts := make([]*entities.Account, 0, len(dtos))
		for _, d := range dtos {
			accounts = append(accounts, dto.FromAccountPersistenceDTO(d))
		}
		return accounts, nil
	}

	// Fallback: try to parse as object/map (old format from CLI)
	var accountMap map[string]interface{}
	if err := json.Unmarshal(data, &accountMap); err != nil {
		return nil, fmt.Errorf("failed to parse accounts file: %w", err)
	}

	// Convert old format to new format
	accounts := make([]*entities.Account, 0, len(accountMap))
	for orgUUID, val := range accountMap {
		accountData, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract OAuth token info
		var accessToken, refreshToken string
		var expiresAt int64
		if oauthToken, ok := accountData["oauth_token"].(map[string]interface{}); ok {
			if at, ok := oauthToken["access_token"].(string); ok {
				accessToken = at
			}
			if rt, ok := oauthToken["refresh_token"].(string); ok {
				refreshToken = rt
			}
			if exp, ok := oauthToken["expires_at"].(float64); ok {
				expiresAt = int64(exp)
			}
		}

		// Extract status
		status := "active"
		if s, ok := accountData["status"].(string); ok {
			status = s
		}

		// Create account entity
		account := &entities.Account{
			ID:               orgUUID,
			Name:             orgUUID,
			OrganizationUUID: orgUUID,
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			ExpiresAt:        time.Unix(expiresAt, 0),
			Status:           entities.AccountStatus(status),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Create creates and persists a new account
func (r *JSONAccountPersistenceRepository) Create(ctx context.Context, account *entities.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing accounts
	accounts, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Check for duplicates
	for _, a := range accounts {
		if strings.EqualFold(a.Name, account.Name) {
			return fmt.Errorf("account with name already exists")
		}
		if account.OrganizationUUID != "" && a.OrganizationUUID == account.OrganizationUUID {
			return fmt.Errorf("account with organization UUID already exists")
		}
	}

	// Add new account
	accounts = append(accounts, account)

	// Save all back to disk
	return r.saveToDisk(accounts)
}

// Update updates and persists an existing account
func (r *JSONAccountPersistenceRepository) Update(ctx context.Context, account *entities.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing accounts
	accounts, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Find and update the account
	found := false
	for i, a := range accounts {
		if a.ID == account.ID {
			accounts[i] = account
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("account not found: %s", account.ID)
	}

	// Save all back to disk
	return r.saveToDisk(accounts)
}

// Delete deletes an account from persistent storage
func (r *JSONAccountPersistenceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load all existing accounts
	accounts, err := r.loadFromDisk()
	if err != nil {
		return err
	}

	// Find and remove the account
	found := false
	for i, a := range accounts {
		if a.ID == id {
			accounts = append(accounts[:i], accounts[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("account not found: %s", id)
	}

	// Save all back to disk
	return r.saveToDisk(accounts)
}

// loadFromDisk loads accounts from disk (internal helper, requires lock)
func (r *JSONAccountPersistenceRepository) loadFromDisk() ([]*entities.Account, error) {
	accountsFile := filepath.Join(r.dataFolder, "accounts.json")

	data, err := os.ReadFile(accountsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entities.Account{}, nil
		}
		return nil, fmt.Errorf("failed to read accounts file: %w", err)
	}

	var dtos []*dto.AccountPersistenceDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return nil, fmt.Errorf("failed to parse accounts file: %w", err)
	}

	accounts := make([]*entities.Account, 0, len(dtos))
	for _, d := range dtos {
		accounts = append(accounts, dto.FromAccountPersistenceDTO(d))
	}

	return accounts, nil
}

// saveToDisk saves accounts to disk (internal helper, requires lock)
func (r *JSONAccountPersistenceRepository) saveToDisk(accounts []*entities.Account) error {
	accountsFile := filepath.Join(r.dataFolder, "accounts.json")

	// Convert entities to DTOs
	dtos := make([]*dto.AccountPersistenceDTO, 0, len(accounts))
	for _, account := range accounts {
		dtos = append(dtos, dto.ToAccountPersistenceDTO(account))
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(dtos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	// Write to temporary file first
	tmpFile := accountsFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write accounts file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, accountsFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename accounts file: %w", err)
	}

	return nil
}
