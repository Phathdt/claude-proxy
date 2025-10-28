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

	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	"github.com/google/uuid"
)

// AccountDTO represents the JSON structure for account persistence
type AccountDTO struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	OrganizationUUID string  `json:"organization_uuid"`
	AccessToken      string  `json:"access_token"`
	RefreshToken     string  `json:"refresh_token"`
	ExpiresAt        string  `json:"expires_at"` // RFC3339/ISO 8601 datetime
	Status           string  `json:"status"`
	RateLimitedUntil *string `json:"rate_limited_until,omitempty"` // RFC3339/ISO 8601 datetime, nil if not rate limited
	LastRefreshError string  `json:"last_refresh_error,omitempty"` // Error message from last refresh attempt
	CreatedAt        string  `json:"created_at"`                   // RFC3339/ISO 8601 datetime
	UpdatedAt        string  `json:"updated_at"`                   // RFC3339/ISO 8601 datetime
}

// JSONAccountRepository implements AccountRepository using JSON file storage
type JSONAccountRepository struct {
	dataFolder string
	accounts   map[string]*entities.Account
	mu         sync.RWMutex
}

// NewJSONAccountRepository creates a new JSON account repository
func NewJSONAccountRepository(dataFolder string) (interfaces.AccountRepository, error) {
	repo := &JSONAccountRepository{
		dataFolder: expandPath(dataFolder),
		accounts:   make(map[string]*entities.Account),
	}

	// Create data folder if it doesn't exist
	if err := os.MkdirAll(repo.dataFolder, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create data folder: %w", err)
	}

	// Load existing accounts
	if err := repo.load(); err != nil {
		return nil, fmt.Errorf("failed to load accounts: %w", err)
	}

	return repo, nil
}

// Create creates a new app account
func (r *JSONAccountRepository) Create(ctx context.Context, account *entities.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if account with same name or organization UUID already exists
	for _, a := range r.accounts {
		if strings.EqualFold(a.Name, account.Name) {
			return fmt.Errorf("account with name already exists")
		}
		if account.OrganizationUUID != "" && a.OrganizationUUID == account.OrganizationUUID {
			return fmt.Errorf("account with organization UUID already exists")
		}
	}

	// Generate ID if not set (should never happen, but defensive)
	if account.ID == "" {
		account.ID = uuid.Must(uuid.NewV7()).String()
	}

	r.accounts[account.ID] = account
	return r.save()
}

// GetByID retrieves an app account by ID
func (r *JSONAccountRepository) GetByID(ctx context.Context, id string) (*entities.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, exists := r.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found: %s", id)
	}

	// Return a copy
	accountCopy := *account
	return &accountCopy, nil
}

// List retrieves all app accounts
func (r *JSONAccountRepository) List(ctx context.Context) ([]*entities.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]*entities.Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		accountCopy := *account
		accounts = append(accounts, &accountCopy)
	}

	return accounts, nil
}

// Update updates an existing app account
func (r *JSONAccountRepository) Update(ctx context.Context, account *entities.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[account.ID]; !exists {
		return fmt.Errorf("account not found: %s", account.ID)
	}

	// Check if name or organization UUID changed and conflicts with another account
	for id, a := range r.accounts {
		if id != account.ID {
			if strings.EqualFold(a.Name, account.Name) {
				return fmt.Errorf("account with name already exists")
			}
			if account.OrganizationUUID != "" && a.OrganizationUUID == account.OrganizationUUID {
				return fmt.Errorf("account with organization UUID already exists")
			}
		}
	}

	r.accounts[account.ID] = account
	return r.save()
}

// Delete deletes an app account by ID
func (r *JSONAccountRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[id]; !exists {
		return fmt.Errorf("account not found: %s", id)
	}

	delete(r.accounts, id)
	return r.save()
}

// GetActiveAccounts retrieves all active app accounts
func (r *JSONAccountRepository) GetActiveAccounts(ctx context.Context) ([]*entities.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var activeAccounts []*entities.Account
	for _, account := range r.accounts {
		if account.IsActive() {
			accountCopy := *account
			activeAccounts = append(activeAccounts, &accountCopy)
		}
	}

	return activeAccounts, nil
}

// load loads accounts from disk
func (r *JSONAccountRepository) load() error {
	accountsFile := filepath.Join(r.dataFolder, "accounts.json")

	data, err := os.ReadFile(accountsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No accounts yet
		}
		return fmt.Errorf("failed to read accounts file: %w", err)
	}

	// Try to parse as array first (new format)
	var dtos []*AccountDTO
	if err := json.Unmarshal(data, &dtos); err == nil && len(dtos) > 0 {
		// Successfully parsed as array
		for _, dto := range dtos {
			account := accountDtoToEntity(dto)
			r.accounts[account.ID] = account
		}
		return nil
	}

	// Fallback: try to parse as object/map (old format from CLI)
	var accountMap map[string]interface{}
	if err := json.Unmarshal(data, &accountMap); err != nil {
		return fmt.Errorf("failed to parse accounts file: %w", err)
	}

	// Convert old format to new format
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

		r.accounts[account.ID] = account
	}

	return nil
}

// save persists accounts to disk atomically
func (r *JSONAccountRepository) save() error {
	accountsFile := filepath.Join(r.dataFolder, "accounts.json")

	// Convert map to slice
	accounts := make([]*AccountDTO, 0, len(r.accounts))
	for _, account := range r.accounts {
		accounts = append(accounts, accountEntityToDTO(account))
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(accounts, "", "  ")
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

// accountEntityToDTO converts entity to DTO
func accountEntityToDTO(account *entities.Account) *AccountDTO {
	dto := &AccountDTO{
		ID:               account.ID,
		Name:             account.Name,
		OrganizationUUID: account.OrganizationUUID,
		AccessToken:      account.AccessToken,
		RefreshToken:     account.RefreshToken,
		ExpiresAt:        account.ExpiresAt.Format(time.RFC3339),
		Status:           string(account.Status),
		LastRefreshError: account.LastRefreshError,
		CreatedAt:        account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        account.UpdatedAt.Format(time.RFC3339),
	}

	// Convert RateLimitedUntil pointer
	if account.RateLimitedUntil != nil {
		timestamp := account.RateLimitedUntil.Format(time.RFC3339)
		dto.RateLimitedUntil = &timestamp
	}

	return dto
}

// accountDtoToEntity converts DTO to entity
func accountDtoToEntity(dto *AccountDTO) *entities.Account {
	expiresAt, _ := time.Parse(time.RFC3339, dto.ExpiresAt)
	createdAt, _ := time.Parse(time.RFC3339, dto.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, dto.UpdatedAt)

	account := &entities.Account{
		ID:               dto.ID,
		Name:             dto.Name,
		OrganizationUUID: dto.OrganizationUUID,
		AccessToken:      dto.AccessToken,
		RefreshToken:     dto.RefreshToken,
		ExpiresAt:        expiresAt,
		Status:           entities.AccountStatus(dto.Status),
		LastRefreshError: dto.LastRefreshError,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}

	// Convert RateLimitedUntil pointer
	if dto.RateLimitedUntil != nil {
		t, _ := time.Parse(time.RFC3339, *dto.RateLimitedUntil)
		account.RateLimitedUntil = &t
	}

	return account
}
