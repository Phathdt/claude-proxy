package token

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Token represents an API token
type Token struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Status      string `json:"status"` // active, inactive
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	UsageCount  int    `json:"usage_count"`
	LastUsedAt  *int64 `json:"last_used_at,omitempty"`
}

// Manager manages API tokens
type Manager struct {
	dataFolder string
	tokens     map[string]*Token
	mu         sync.RWMutex
}

// NewManager creates a new token manager
func NewManager(dataFolder string) *Manager {
	return &Manager{
		dataFolder: expandPath(dataFolder),
		tokens:     make(map[string]*Token),
	}
}

// Initialize loads existing tokens
func (m *Manager) Initialize() error {
	tokensFile := filepath.Join(m.dataFolder, "tokens.json")

	// Create data folder if it doesn't exist
	if err := os.MkdirAll(m.dataFolder, 0700); err != nil {
		return fmt.Errorf("failed to create data folder: %w", err)
	}

	// Load existing tokens if file exists
	if _, err := os.Stat(tokensFile); err == nil {
		data, err := os.ReadFile(tokensFile)
		if err != nil {
			return fmt.Errorf("failed to read tokens file: %w", err)
		}

		var tokens []*Token
		if err := json.Unmarshal(data, &tokens); err != nil {
			return fmt.Errorf("failed to parse tokens file: %w", err)
		}

		m.mu.Lock()
		for _, token := range tokens {
			m.tokens[token.ID] = token
		}
		m.mu.Unlock()
	}

	return nil
}

// save persists tokens to disk
func (m *Manager) save() error {
	tokensFile := filepath.Join(m.dataFolder, "tokens.json")

	m.mu.RLock()
	tokens := make([]*Token, 0, len(m.tokens))
	for _, token := range m.tokens {
		tokens = append(tokens, token)
	}
	m.mu.RUnlock()

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	if err := os.WriteFile(tokensFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	return nil
}

// GenerateKey generates a random API key
func GenerateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "tk_" + hex.EncodeToString(bytes), nil
}

// CreateToken creates a new token
func (m *Manager) CreateToken(name, key, status string) (*Token, error) {
	if name == "" {
		return nil, fmt.Errorf("token name is required")
	}
	if key == "" {
		return nil, fmt.Errorf("token key is required")
	}

	// Check if key already exists
	m.mu.RLock()
	for _, t := range m.tokens {
		if t.Key == key {
			m.mu.RUnlock()
			return nil, fmt.Errorf("token key already exists")
		}
	}
	m.mu.RUnlock()

	token := &Token{
		ID:         generateID(),
		Name:       name,
		Key:        key,
		Status:     status,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
		UsageCount: 0,
	}

	m.mu.Lock()
	m.tokens[token.ID] = token
	m.mu.Unlock()

	if err := m.save(); err != nil {
		return nil, err
	}

	return token, nil
}

// ListTokens returns all tokens
func (m *Manager) ListTokens() []*Token {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokens := make([]*Token, 0, len(m.tokens))
	for _, token := range m.tokens {
		tokens = append(tokens, token)
	}
	return tokens
}

// GetToken returns a token by ID
func (m *Manager) GetToken(id string) (*Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	token, exists := m.tokens[id]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	return token, nil
}

// UpdateToken updates a token
func (m *Manager) UpdateToken(id, name, key, status string) (*Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, exists := m.tokens[id]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	// Check if key changed and if it conflicts with another token
	if key != token.Key {
		for _, t := range m.tokens {
			if t.ID != id && t.Key == key {
				return nil, fmt.Errorf("token key already exists")
			}
		}
	}

	token.Name = name
	token.Key = key
	token.Status = status
	token.UpdatedAt = time.Now().Unix()

	if err := m.save(); err != nil {
		return nil, err
	}

	return token, nil
}

// DeleteToken deletes a token
func (m *Manager) DeleteToken(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tokens[id]; !exists {
		return fmt.Errorf("token not found")
	}

	delete(m.tokens, id)

	return m.save()
}

// ValidateToken validates a token and increments usage
func (m *Manager) ValidateToken(key string) (*Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, token := range m.tokens {
		if token.Key == key {
			if token.Status != "active" {
				return nil, fmt.Errorf("token is inactive")
			}

			// Increment usage
			token.UsageCount++
			now := time.Now().Unix()
			token.LastUsedAt = &now

			if err := m.save(); err != nil {
				return nil, err
			}

			return token, nil
		}
	}

	return nil, fmt.Errorf("invalid token")
}

// generateID generates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
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
