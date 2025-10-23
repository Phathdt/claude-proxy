# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`claude-proxy` is a Go-based API service designed for wallet risk assessment and proxy functionality. Built with:
- **Framework**: Gin HTTP framework with Uber FX for dependency injection
- **Database**: PostgreSQL with pgx driver
- **Configuration**: Viper with YAML config + env variables
- **CLI**: urfave/cli v2 for command-line interface
- **Logging**: Custom structured logging via service-context package

## Architecture

### Application Structure

```
claude-proxy/
├── cmd/api/          # API server setup and providers
│   ├── server.go     # HTTP server lifecycle management
│   └── providers.go  # Dependency injection providers
├── cli/              # CLI command definitions
├── config/           # Configuration management
├── pkg/              # Shared packages
│   ├── errors/       # Custom error types with HTTP context
│   ├── telegram/     # Telegram notification client
│   └── address/      # Address validation utilities
└── main.go          # Application entry point
```

### Dependency Injection Architecture

The application uses **Uber FX** for dependency injection with a provider-based architecture:

- **CoreProviders**: Configuration, ServiceContext, Logger, Database connection
- **WalletCheckerProviders**: Domain-specific providers (Telegram client, etc.)
- **APIProviders**: Combines core and domain providers with Gin engine

Providers are defined in `cmd/api/providers.go:21-46`. Each provider function returns dependencies that FX automatically wires together.

### Configuration System

Configuration uses **Viper** with YAML files and environment variable overrides:

- Config file: `config.yaml` (see `config.example.yaml` for template)
- Environment variables: Use double underscore for nested keys (e.g., `server__port=4000`)
- Structure defined in `config/config.go:12-74`
- Key sections: `database`, `server`, `logger`, `telegram`, `wallet_checker`

### Error Handling

Custom error system with HTTP-aware panic recovery (`pkg/errors/app_error.go`):

- **AppError interface**: All app errors implement `StatusCode()`, `ErrorCode()`, `Message()`, `Details()`
- **Panic-based flow**: Handlers can `panic(appError)` - middleware catches and converts to proper JSON response
- **Error constructors**: `NewBadRequestError()`, `NewNotFoundError()`, `NewValidationError()`, etc.
- **Recovery middleware**: Configured in `cmd/api/providers.go:90-135`

### Service Context Pattern

Uses `github.com/phathdt/service-context` for component lifecycle:

- **ServiceContext**: Central registry for components (database, cache, etc.)
- **Component Loading**: Components auto-load via `sc.Load()` in `cmd/api/providers.go:54-74`
- **Database Access**: Retrieve with `sc.MustGet("postgres").(pgxc.PgxComp).GetConn()`

## Commands

### Development

```bash
# Run application (default with config.yaml)
make run
go run .

# Run with custom config
go run . server --config custom.yaml
go run . api --config custom.yaml

# Build binary
make build  # outputs to bin/claude-proxy

# Install dependencies
make deps
go mod tidy
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

### Docker & Database

```bash
# Start services (if docker-compose exists)
make docker-up
docker-compose up -d

# Stop services
make docker-down

# Generate sqlc code (if sqlc.yaml exists)
make sqlc-generate
```

### Development Workflow

```bash
# Setup development environment
make dev-setup  # formats code and tidies dependencies

# Run with Docker services
make dev  # starts docker-compose, waits 5s, then runs app
```

## Development Guidelines

### Adding New API Endpoints

1. Define handler in appropriate module under `cmd/api/` or create new module
2. Register routes in `cmd/api/server.go` or module-specific router
3. Use structured logging via `sctx.GlobalLogger().GetLogger("component-name")`
4. Return errors by panicking with AppError: `panic(errors.NewBadRequestError(...))`

### Testing Strategy

- Unit tests: Mock external dependencies, test business logic in isolation
- Integration tests: Test with real database connections
- Use `_test.go` suffix for all test files
- Example: `pkg/address/validator_test.go`

### Configuration Changes

1. Update structs in `config/config.go`
2. Add defaults in `LoadConfig()` if necessary (e.g., logger defaults at line 105-110)
3. Update `config.example.yaml` template
4. Document env variable names using double underscore convention

### Logging Best Practices

```go
logger := sctx.GlobalLogger().GetLogger("component-name")
logger.Info("message")
logger.Withs(sctx.Fields{"key": "value"}).Error("error message")
```

Structured fields should use `sctx.Fields` map for consistent JSON logging.

### Database Patterns

- Connection pool managed by service-context pgxc component
- Access via `db := sc.MustGet("postgres").(pgxc.PgxComp).GetConn()`
- Use context-aware queries: `db.QueryRow(ctx, "SELECT ...")`
- Consider using sqlc for type-safe SQL queries (sqlc-generate target exists)

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
export server__port=8080
export database__uri="postgres://..."
export telegram__enabled=true
export logger__level=debug
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
