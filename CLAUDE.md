# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`claude-proxy` is a full-stack application with a Go backend API and React frontend admin dashboard, designed for wallet risk assessment and proxy functionality.

### Backend Stack
- **Framework**: Gin HTTP framework with Uber FX for dependency injection
- **Configuration**: Viper with YAML config + env variables
- **CLI**: urfave/cli v2 for command-line interface
- **Logging**: Custom structured logging via service-context package

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

### Application Structure

```
claude-proxy/
├── cmd/api/          # Backend API server setup and providers
│   ├── server.go     # HTTP server lifecycle, embedded frontend serving
│   └── providers.go  # Dependency injection providers
├── cli/              # CLI command definitions
├── config/           # Configuration management
├── pkg/              # Shared packages
│   ├── errors/       # Custom error types with HTTP context
│   ├── telegram/     # Telegram notification client
│   └── address/      # Address validation utilities
├── frontend/         # React TypeScript frontend
│   ├── src/
│   │   ├── components/  # React components (app-tokens, layout, ui)
│   │   ├── hooks/       # React Query hooks
│   │   ├── lib/         # API client, utilities, mock data
│   │   ├── pages/       # Page components (login, dashboard, tokens, app-tokens)
│   │   └── types/       # TypeScript type definitions
│   ├── index.html       # Entry HTML
│   ├── vite.config.ts   # Vite configuration with proxy setup
│   └── package.json     # Frontend dependencies
└── main.go             # Application entry point, embeds frontend
```

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

- Config file: `config.yaml` (see `config.example.yaml` for template)
- Environment variables: Uppercase with double underscore for nested keys (e.g., `SERVER__PORT=4000`)
- Structure defined in `config/config.go`
- Key sections: `server`, `logger`, `telegram`, `wallet_checker`

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

**Backend:**
```bash
# Run backend only (port 4000)
make be
go run .

# Run with custom config
go run . server --config custom.yaml
```

**Frontend:**
```bash
# Install frontend dependencies
cd frontend && pnpm install

# Run frontend dev server (port 5173)
make fe
cd frontend && pnpm dev

# Lint frontend
cd frontend && pnpm lint
cd frontend && pnpm lint:fix

# Format frontend code
cd frontend && pnpm format
cd frontend && pnpm format:check
```

**Full Stack:**
```bash
# Run both backend and frontend (in separate terminals)
make be  # Terminal 1
make fe  # Terminal 2

# Build production binary with embedded frontend
make build  # outputs to bin/claude-proxy
```

**Dependencies:**
```bash
# Backend dependencies
make deps
go mod tidy

# Frontend dependencies (already in package.json)
cd frontend && pnpm install
```

### Testing

```bash
# Run all tests
make test
go test ./... -v

# Run only unit tests
make test-unit
go test ./modules/... -v -short

# Run integration tests
make test-integration
go test ./... -v -run Integration

# Generate coverage report
make test-coverage  # creates coverage.html

# Watch mode (requires entr)
make test-watch
```

### Code Formatting

```bash
# Format all Go files (runs gofmt, goimports, golines, gofumpt)
make format

# Check formatting without changes
make format-check
```

### Docker

```bash
# Start services (if docker-compose exists)
make docker-up
docker-compose up -d

# Stop services
make docker-down
```

### Development Workflow

```bash
# Setup development environment
make dev-setup  # formats code and tidies dependencies

# Run with Docker services
make dev  # starts docker-compose, waits 5s, then runs app
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
1. Define handler in appropriate module under `cmd/api/` or create new module
2. Register routes in `cmd/api/server.go` under the `/api` group
3. Use structured logging via `sctx.GlobalLogger().GetLogger("component-name")`
4. Return errors by panicking with AppError: `panic(errors.NewBadRequestError(...))`

### Testing Strategy

- Unit tests: Mock external dependencies, test business logic in isolation
- Use `_test.go` suffix for all test files
- Example: `pkg/address/validator_test.go`

### Configuration Changes

1. Update structs in `config/config.go`
2. Add defaults in `LoadConfig()` if necessary
3. Update `config.example.yaml` template
4. Document env variable names using double underscore convention

### Logging Best Practices

```go
logger := sctx.GlobalLogger().GetLogger("component-name")
logger.Info("message")
logger.Withs(sctx.Fields{"key": "value"}).Error("error message")
```

Structured fields should use `sctx.Fields` map for consistent JSON logging.

## Key Technical Details

### CLI Commands

- **Default**: Runs server with `config.yaml`
- **server/s**: Start API server (alias for default)
- **api/a**: Same as server (backward compatibility)

All commands defined in `main.go:12-50`, implemented in `cli/cli.go`.

### HTTP Server Features

- **Middleware stack**: Logger → Recovery → CORS → Timeout (30s)
- **CORS**: Allows all origins (`Access-Control-Allow-Origin: *`)
- **Health check**: `GET /health` returns status and timestamp
- **Gin release mode**: Production-optimized, minimal logging

### Telegram Integration

Optional notification system (`pkg/telegram/telegram.go`):
- Enable via `telegram.enabled: true` in config
- Supports Markdown formatting
- Configurable timeout and retry logic
- Used for monitoring and alerts

### Environment Variables

Override any YAML config with env vars:
```bash
# Examples
export SERVER__PORT=8080
export TELEGRAM__ENABLED=true
export LOGGER__LEVEL=debug
```

## Common Patterns

### Adding a New Domain Module

1. Create package under project root or `pkg/`
2. Define models, business logic, handlers
3. Create provider functions in module or `cmd/api/providers.go`
4. Add to provider chain in `cmd/api/providers.go:APIProviders`
5. Register routes in server startup

### Middleware Pattern

Add middleware in `NewGinEngine()` at `cmd/api/providers.go:82-161`:
```go
engine.Use(func(c *gin.Context) {
    // Your middleware logic
    c.Next()
})
```

### Component Registration

Add to service context in `InitServiceContext()`:
```go
sc.Load() // Auto-discovers components from config
// Or manually register components as needed
```

## Admin Dashboard Features

The frontend provides an admin interface for managing:

1. **Authentication** (`/login`):
   - Mock authentication (currently accepts any email/password)
   - Stores token in localStorage
   - Protected route wrapper for admin pages

2. **Dashboard** (`/admin/dashboard`):
   - Overview and statistics
   - Quick access to main features

3. **Tokens Management** (`/admin/tokens`):
   - CRUD operations for API tokens
   - Token status (active/inactive)
   - Usage tracking (count, last used)

4. **App Tokens Management** (`/admin/app-tokens`):
   - OAuth application management
   - Fields: name, email, orgId, type (oauth/cookies), accountType (pro/max)
   - No OAuth implementation details (clientId, secret, etc. removed)
   - Status and usage tracking

**Current State:** Frontend uses mock data in `frontend/src/lib/api.ts`. Replace mock functions with real API calls when backend endpoints are ready.
