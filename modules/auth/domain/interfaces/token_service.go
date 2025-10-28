package interfaces

import (
	"context"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/entities"

	"github.com/phathdt/service-context/core"
)

// TokenService defines the interface for token management operations
type TokenService interface {
	// CreateToken creates a new token
	CreateToken(
		ctx context.Context,
		name, key string,
		status entities.TokenStatus,
		role entities.TokenRole,
	) (*entities.Token, error)

	// GetTokenByID retrieves a token by ID
	GetTokenByID(ctx context.Context, id string) (*entities.Token, error)

	// GetTokenByKey retrieves a token by its key
	GetTokenByKey(ctx context.Context, key string) (*entities.Token, error)

	// ListTokens retrieves tokens with optional filtering and pagination
	// Pagination metadata is injected into the paging pointer
	ListTokens(ctx context.Context, query *dto.TokenQueryParams, paging *core.Paging) ([]*entities.Token, error)

	// UpdateToken updates an existing token
	UpdateToken(
		ctx context.Context,
		id, name, key string,
		status entities.TokenStatus,
		role entities.TokenRole,
	) (*entities.Token, error)

	// DeleteToken deletes a token by ID
	DeleteToken(ctx context.Context, id string) error

	// ValidateToken validates a token key and returns the token if valid
	ValidateToken(ctx context.Context, key string) (*entities.Token, error)

	// Sync syncs in-memory data to persistent storage
	Sync(ctx context.Context) error

	// FinalSync performs final sync on shutdown
	FinalSync(ctx context.Context) error
}
