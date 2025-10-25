package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"
)

// AccountDTO represents the JSON structure for account persistence
type AccountDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	OrganizationUUID string `json:"organization_uuid"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresAt        int64  `json:"expires_at"`
	Status           string `json:"status"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
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

	// Generate ID if not set
	if account.ID == "" {
		account.ID = generateAccountID()
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

	var dtos []*AccountDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return fmt.Errorf("failed to parse accounts file: %w", err)
	}

	for _, dto := range dtos {
		account := accountDtoToEntity(dto)
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
	return &AccountDTO{
		ID:               account.ID,
		Name:             account.Name,
		OrganizationUUID: account.OrganizationUUID,
		AccessToken:      account.AccessToken,
		RefreshToken:     account.RefreshToken,
		ExpiresAt:        account.ExpiresAt.Unix(),
		Status:           string(account.Status),
		CreatedAt:        account.CreatedAt.Unix(),
		UpdatedAt:        account.UpdatedAt.Unix(),
	}
}

// accountDtoToEntity converts DTO to entity
func accountDtoToEntity(dto *AccountDTO) *entities.Account {
	return &entities.Account{
		ID:               dto.ID,
		Name:             dto.Name,
		OrganizationUUID: dto.OrganizationUUID,
		AccessToken:      dto.AccessToken,
		RefreshToken:     dto.RefreshToken,
		ExpiresAt:        time.Unix(dto.ExpiresAt, 0),
		Status:           entities.AccountStatus(dto.Status),
		CreatedAt:        time.Unix(dto.CreatedAt, 0),
		UpdatedAt:        time.Unix(dto.UpdatedAt, 0),
	}
}

// generateAccountID generates a unique ID for an account
func generateAccountID() string {
	return fmt.Sprintf("app_%d", time.Now().UnixNano())
}
