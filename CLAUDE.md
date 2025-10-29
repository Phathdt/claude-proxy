# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`claude-proxy` is a full-stack OAuth 2.0-based Claude API reverse proxy with multi-account support, automatic token refresh, and admin dashboard. It enables secure, scalable access to Claude API with built-in account management and load balancing.

### Core Purpose
- **OAuth 2.0 Proxy**: Authenticate users via Claude OAuth, proxy requests to Claude API
- **Multi-Account Support**: Load balance across multiple Claude accounts
- **Automatic Token Refresh**: Hourly cronjob + on-demand refresh with 60-second buffer before expiration
- **Admin Dashboard**: React-based UI for account management and OAuth setup

### Backend Stack
- **Framework**: Gin HTTP framework with Uber FX for dependency injection
- **Configuration**: Viper with YAML config + env variables
- **CLI**: urfave/cli v2 for command-line interface
- **Logging**: Structured logging via service-context package
- **Scheduling**: robfig/cron/v3 for in-memory job scheduling (no external queue)

### Frontend Stack
- **Framework**: React 19 with TypeScript
- **Build Tool**: Vite 7
- **Routing**: React Router DOM v7
- **State Management**: TanStack React Query v5
- **Styling**: TailwindCSS v4 with shadcn/ui components
- **Icons**: Lucide React
- **Linting**: ESLint v9 with Prettier integration
- **Package Manager**: pnpm

## Architecture

### Domain-Driven Design Structure

The backend follows **Domain-Driven Design (DDD)** pattern with clear separation of concerns:

```
modules/proxy/
├── domain/               # Domain layer - business rules & interfaces
│   ├── entities/         # Account, Token entities with business logic
│   └── interfaces/       # Service & repository interfaces
├── application/          # Application layer - business logic orchestration
│   ├── services/         # AccountService, TokenService, ProxyService
│   └── dto/             # Data Transfer Objects (request/response models)
└── infrastructure/       # Infrastructure layer - external integrations
    ├── clients/         # ClaudeAPIClient for proxying requests
    ├── repositories/    # JSONAccountRepository, JSONTokenRepository
    └── jobs/           # Scheduler for automatic token refresh
```

**Key Design Principles**:
- **Dependency Inversion**: High-level modules depend on abstractions (interfaces)
- **Repository Pattern**: Data persistence abstraction (JSON files)
- **Service Layer**: Orchestrates domain logic and repositories
- **Entity Aggregate**: Account entity manages itself and associated tokens

### Full Project Structure

```
claude-proxy/
├── cmd/api/              # Server entry point & DI setup
│   ├── server.go        # HTTP server, routes, middleware
│   ├── providers.go      # Uber FX dependency injection
│   └── handlers/         # HTTP request handlers
├── config/               # Configuration management (Viper)
├── pkg/                  # Shared packages
│   ├── errors/          # Custom AppError with HTTP context
│   └── telegram/        # Optional Telegram notifications
├── modules/             # Domain modules (see DDD structure above)
├── cli/                 # CLI command implementations
├── frontend/            # React TypeScript admin dashboard
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── hooks/       # React Query hooks
│   │   ├── lib/         # API client, utilities
│   │   ├── pages/       # Page components
│   │   └── types/       # TypeScript type definitions
│   └── vite.config.ts   # Vite with backend proxy setup
└── main.go              # Entry point (embeds frontend)
```

### Token Refresh Flow

There are **two triggers** for token refresh:

1. **Cronjob Trigger** (Every Hour): `modules/proxy/infrastructure/jobs/scheduler.go`
   - Runs at minute 0 of every hour (`0 * * * *` cron expression)
   - Iterates all active accounts, checks `NeedsRefresh()` (60s buffer)
   - Calls `AccountService.GetValidToken()` → `refreshToken()` → `TokenRefresher.RefreshAccessToken()` (OAuth API)
   - Logs summary (refreshed/failed/skipped counts)
   - 5-minute timeout for entire job execution

2. **API Request Trigger** (On-Demand): `modules/proxy/application/services/proxy_service.go`
   - When user sends request to `/api/proxy/*`
   - Calls `GetValidAccount()` for load-balanced account selection
   - If token within 60s of expiration, automatically refreshes before using
   - Transparent to user - refresh happens inside `GetValidToken()`

**Account Refresh Logic** (in `modules/proxy/domain/entities/account.go`):
```go
func (a *Account) NeedsRefresh() bool {
    return time.Now().After(a.ExpiresAt.Add(-60 * time.Second))
}
```
- Returns `true` if current time > (expiration - 60 seconds)
- Both triggers use same logic for consistency

### Load Balancing Strategy

From `proxy_service.go`:
- **Round-Robin Selection**: Uses timestamp modulo to distribute requests
- **Health Filtering**: Prioritizes accounts that don't need immediate refresh
- **No Persistent State**: Selection algorithm is stateless
- **All Accounts Eligible**: Load balancing works across all active accounts

### Frontend-Backend Integration

The application uses **Go embed** to bundle the React frontend into a single binary:

1. **Development Mode**: Frontend runs on port 5173 (Vite dev server), backend on port 4000
   - Vite proxy configured to forward `/api/*` requests to backend
   - Hot reload for frontend development

2. **Production Build**: Frontend compiled and embedded into Go binary
   - `make build` runs `pnpm build` then `go build`
   - Embedded files accessed via `embed.FS` in `main.go`
   - Backend serves static files with SPA routing support
   - All requests to `/api/*` handled by backend, others serve frontend

3. **API Routing**:
   - Backend API routes prefixed with `/api` (e.g., `/api/health`)
   - Frontend routes handled by React Router (client-side)
   - 404s from static serving fallback to `index.html` for SPA

### Dependency Injection Architecture

The application uses **Uber FX** for dependency injection with a provider-based architecture:

- **CoreProviders**: Configuration, ServiceContext, Logger
- **WalletCheckerProviders**: Domain-specific providers (Telegram client, etc.)
- **APIProviders**: Combines core and domain providers with Gin engine

Providers are defined in `cmd/api/providers.go`. Each provider function returns dependencies that FX automatically wires together.

### Configuration System

Configuration uses **Viper** with YAML files and environment variable overrides:

- **Config File**: `config.yaml` (see `config.example.yaml` for template)
- **Environment Variables**: Uppercase with double underscore for nested keys (e.g., `SERVER__PORT=4000`)
- **Structure**: Defined in `config/config.go`
- **Key Sections**:
  - `server`: Host/port (default: 0.0.0.0:4000)
  - `oauth`: Client ID, authorize/token URLs, redirect URI, scopes
  - `claude`: Claude API base URL
  - `storage`: Data folder for JSON persistence (default: `~/.claude-proxy/data`)
  - `logger`: Level, format (text/json)
  - `retry`: Max retries, retry delay
  - `telegram`: Optional notification alerts

**Data Storage**: JSON files in `~/.claude-proxy/data/`
- `account.json`: Persisted account credentials with tokens
- Per-account file structure: ID, name, tokens, expiration times, OAuth state

### Error Handling

Custom error system with HTTP-aware panic recovery (`pkg/errors/app_error.go`):

- **AppError interface**: All app errors implement `StatusCode()`, `ErrorCode()`, `Message()`, `Details()`
- **Panic-based flow**: Handlers can `panic(appError)` - middleware catches and converts to proper JSON response
- **Error constructors**: `NewBadRequestError()`, `NewNotFoundError()`, `NewValidationError()`, etc.
- **Recovery middleware**: Configured in `cmd/api/providers.go:90-135`

### Service Context Pattern

Uses `github.com/phathdt/service-context` for component lifecycle:

- **ServiceContext**: Central registry for components (logger, cache, etc.)
- **Component Loading**: Components auto-load via `sc.Load()` in `cmd/api/providers.go`
- **Global Logger**: Access via `sctx.GlobalLogger().GetLogger("component-name")`

## Commands

### Development

**Backend (Port 4000):**
```bash
# Run directly
go run . server

# Or with make
make be

# With custom config file
go run . server --config custom.yaml
```

**Frontend (Port 5173):**
```bash
# Development server with hot reload
cd frontend && pnpm dev
# Or: make fe

# Lint and fix
cd frontend && pnpm lint:fix

# Format code
cd frontend && pnpm format
```

**Full Stack Development** (Two Terminals):
```bash
# Terminal 1: Backend
make be
# or: go run . server

# Terminal 2: Frontend
make fe
# or: cd frontend && pnpm dev
```

**Production Build:**
```bash
# Build frontend (bundles React into dist/)
cd frontend && pnpm build

# Build Go binary with embedded frontend (outputs to bin/claude-proxy)
go build -o bin/claude-proxy

# Or use make
make build
```

**Dependencies:**
```bash
# Go dependencies
go mod tidy

# Frontend dependencies
cd frontend && pnpm install
```

### Testing

```bash
# All Go tests
make test
go test ./... -v

# Specific module
go test ./modules/proxy/... -v

# Single test function
go test ./modules/proxy/... -v -run TestFunctionName

# With coverage
go test ./... -cover
go test ./... -coverprofile=coverage.out
```

### Code Formatting

```bash
# Format Go code
make format   # Full format pipeline (gofmt, goimports, etc.)

# Check without modifying
make format-check

# Frontend formatting
cd frontend && pnpm format
cd frontend && pnpm format:check
```

## Development Guidelines

### Frontend Development

**React Query Pattern:**
- API calls in `frontend/src/lib/api.ts` (currently mock data)
- Custom hooks in `frontend/src/hooks/` using React Query
- Example: `useAppTokens()`, `useCreateAppToken()`, `useUpdateAppToken()`

**Component Structure:**
- Pages in `frontend/src/pages/` (login, dashboard, tokens, app-tokens)
- Reusable components in `frontend/src/components/`
- UI primitives in `frontend/src/components/ui/` (shadcn/ui)
- Layout components in `frontend/src/components/layout/`

**Type Safety:**
- All types defined in `frontend/src/types/`
- Use proper TypeScript types, avoid `any`
- DTOs separate from entity types (e.g., `CreateAppTokenDto` vs `AppToken`)

**Styling:**
- TailwindCSS v4 with `@import "tailwindcss"` syntax
- CSS variables for theming in `frontend/src/index.css`
- Prettier plugin automatically orders Tailwind classes

**Routing:**
- Protected routes use `ProtectedRoute` wrapper checking `localStorage.getItem('auth_token')`
- Admin layout at `/admin/*` with sidebar navigation
- SPA routing handled by React Router, backend serves `index.html` for unmatched routes

### Backend Development

**Adding New API Endpoints:**
1. Create handler in `cmd/api/handlers/` if needed
2. Register routes in `cmd/api/server.go` under `/api` group
3. Use structured logging: `sctx.GlobalLogger().GetLogger("component-name")`
4. Panic-based error handling: `panic(errors.NewBadRequestError(...))`

**Working with OAuth & Token Refresh:**
- OAuth client: `modules/auth/infrastructure/clients/oauth_client.go` - Handles PKCE, token exchange, refresh
- OAuth interface: `modules/auth/domain/interfaces/oauth_client.go` - OAuth client interface
- Account management: `modules/auth/application/services/account_service.go`
  - `GetValidToken(ctx, accountID)` - Returns valid token, auto-refreshes if needed
  - `refreshToken(ctx, account)` - Refreshes access token from refresh token
- Token refresh triggers:
  - **Automatic**: Scheduler at `modules/proxy/infrastructure/jobs/scheduler.go` (every hour)
  - **On-Demand**: Inside `GetValidToken()` when token within 60s of expiration

**Account Entity & Persistence:**
- Entity: `modules/proxy/domain/entities/account.go`
  - Methods: `NeedsRefresh()`, `UpdateTokens()`, `UpdateRefreshError()`, `Deactivate()`
- Repository: `modules/proxy/infrastructure/repositories/json_account_repository.go`
  - Reads/writes `~/.claude-proxy/data/account.json`
  - Thread-safe JSON persistence

**Multi-Account Load Balancing:**
- In `proxy_service.go`: `GetValidAccount()` uses round-robin with health filtering
- Algorithm: `time.Now().UnixNano() % len(accounts)` (stateless, no persistent counter)
- Prefers accounts not needing immediate refresh for better UX

### Testing Strategy

- Unit tests: Mock repositories and external services
- Integration tests: Test with actual JSON file storage
- Use `_test.go` suffix for all test files
- Mock OAuth responses when testing token refresh

### Configuration Changes

1. Update struct in `config/config.go`
2. Add defaults in `LoadConfig()` function
3. Update `config.example.yaml` template
4. Document env variable names (uppercase with `__` for nesting)

**Example**: OAuth client ID can be set via `OAUTH__CLIENT_ID=xxx`

### Logging Best Practices

```go
logger := sctx.GlobalLogger().GetLogger("component-name")
logger.Info("starting token refresh")
logger.Withs(sctx.Fields{
    "account_id": accountID,
    "expires_in": expiresIn,
}).Info("token refreshed successfully")
```

Use `sctx.Fields` map for structured, JSON-serializable logging.

## Key Technical Details

### CLI Commands

Defined in `main.go`, implemented in `cli/cli.go`:

- **`claude-proxy server`** (or `claude-proxy`): Start API server with `config.yaml`
- **`--config FILE`**: Specify custom config file path (default: `config.yaml`)

Server startup:
1. Loads configuration from YAML + environment overrides
2. Initializes ServiceContext and logging
3. Creates Gin engine with middleware stack
4. Registers OAuth handlers, proxy handlers, account handlers
5. Starts token refresh scheduler (cronjob)
6. Serves embedded React frontend for SPA routing

### HTTP Server Architecture

**Middleware Stack** (in `cmd/api/providers.go:NewGinEngine()`):
1. **Logger**: Structured request logging with latency
2. **Recovery**: Panic recovery → AppError → JSON response
3. **CORS**: Allow all origins with wildcard headers
4. **Timeout**: 30-second request timeout
5. **SPA Support**: Fallback to `index.html` for undefined routes

**Key Routes** (in `cmd/api/server.go`):
- `GET /health` - Health status
- `GET /oauth/authorize` - Get OAuth authorization URL + PKCE state
- `POST /oauth/exchange` - Exchange auth code for tokens
- `GET /api/accounts` - List accounts
- `POST /api/accounts` - Create account from OAuth code
- `GET /api/proxy/*` - Proxy requests to Claude API
- `GET /api/tokens` - List API tokens (admin)
- Static files: `/*` - Serve React frontend (SPA)

### Token Refresh Scheduler

`modules/proxy/infrastructure/jobs/scheduler.go`:
- Lifecycle managed by Uber FX with `fx.Lifecycle` hooks
- Starts automatically when app starts
- Gracefully stops on shutdown
- Thread-safe cron execution with mutex
- 5-minute timeout for job execution to prevent hangs

### Environment Variables

Override any YAML config using uppercase with double underscores:

```bash
# Server
export SERVER__HOST=127.0.0.1
export SERVER__PORT=8080

# OAuth
export OAUTH__CLIENT_ID=your-client-id
export OAUTH__TOKEN_URL=https://api.claude.ai/oauth/token

# Storage
export STORAGE__DATA_FOLDER=~/.claude-proxy/data

# Logger
export LOGGER__LEVEL=debug
export LOGGER__FORMAT=json

# Telegram (optional)
export TELEGRAM__ENABLED=true
export TELEGRAM__BOT_TOKEN=your-token
export TELEGRAM__CHAT_ID=your-chat-id
```

## Common Patterns

### Adding a New Service

1. Define interface in `modules/proxy/domain/interfaces/`
2. Implement in `modules/proxy/application/services/`
3. Create provider function in `cmd/api/providers.go`
4. Add to `APIProviders` option list
5. Inject as dependency into handlers or other services

**Example** (Token Service):
```go
// Step 1: Interface
type TokenService interface {
    GetTokens(ctx context.Context) ([]Token, error)
}

// Step 2: Implementation
type tokenService struct {
    repo interfaces.TokenRepository
}

// Step 3 & 4: Provider
func NewTokenService(tokenRepo interfaces.TokenRepository) interfaces.TokenService {
    return services.NewTokenService(tokenRepo)
}
```

### Adding a New HTTP Handler

1. Create handler in `cmd/api/handlers/`
2. Implement `func (h *Handler) HandleRequest(c *gin.Context)`
3. Create provider function in `cmd/api/providers.go`
4. Register routes in `cmd/api/server.go` (typically under `/api` group)

**Example Pattern**:
```go
func (h *MyHandler) HandleRequest(c *gin.Context) {
    // Parse request
    // Call service
    // Return JSON or panic with AppError
    c.JSON(http.StatusOK, gin.H{"data": result})
}
```

### Working with DDD Interfaces

Keep high-level code depending on abstractions:
```go
// Good - depends on interface
func NewService(repo interfaces.AccountRepository) *Service {
    return &Service{repo: repo}
}

// Implementation can be swapped (JSON, SQL, etc.)
repo := repositories.NewJSONAccountRepository(dataFolder)
```

### Middleware Pattern

Add middleware in `NewGinEngine()` (`cmd/api/providers.go:NewGinEngine()`):
```go
engine.Use(func(c *gin.Context) {
    // Before request
    c.Set("custom_key", "value")

    c.Next()

    // After request
    statusCode := c.Writer.Status()
})
```

## Admin Dashboard Features

React-based UI for OAuth setup and account management:

1. **Login** (`/login`):
   - Simple authentication (mock or real based on backend)
   - Stores auth token in localStorage
   - Protected routes check `localStorage.getItem('auth_token')`

2. **OAuth Setup** (`/admin/accounts` or similar):
   - Call `GET /oauth/authorize` to get authorization URL
   - User visits authorization URL
   - After OAuth redirect, exchange code via `POST /oauth/exchange`
   - Account auto-created and tokens persisted

3. **Account Management**:
   - View active accounts and their token expiration
   - Monitor refresh status and errors
   - Show account health (active/inactive status)

4. **Token Management** (Admin):
   - API token CRUD operations
   - Usage tracking and revocation

**Current Implementation**:
- Frontend in `frontend/src/pages/` and `frontend/src/components/`
- Backend provides REST API in `cmd/api/handlers/`
- Real API calls via React Query hooks in `frontend/src/hooks/`
