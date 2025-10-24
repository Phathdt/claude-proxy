# Claude Proxy

A full-stack application with a Go backend API and React TypeScript frontend admin dashboard for wallet risk assessment and proxy functionality.

## Tech Stack

### Backend
- **Framework**: Gin HTTP framework with Uber FX dependency injection
- **Configuration**: Viper (YAML + environment variables)
- **CLI**: urfave/cli v2
- **Logging**: Structured logging via service-context

### Frontend
- **Framework**: React 19 + TypeScript
- **Build Tool**: Vite 7
- **Routing**: React Router DOM v7
- **State**: TanStack React Query v5
- **Styling**: TailwindCSS v4 + shadcn/ui
- **Package Manager**: pnpm

## Prerequisites

- Go 1.21+
- Node.js 18+
- pnpm

## Quick Start

### Development

**Install Dependencies:**
```bash
# Backend
go mod tidy

# Frontend
cd frontend && pnpm install
```

**Run Application:**
```bash
# Option 1: Run backend and frontend separately (recommended for development)
make be   # Terminal 1 - Backend on port 4000
make fe   # Terminal 2 - Frontend on port 5173

# Option 2: Run backend only (if you only need the API)
go run .
```

The frontend dev server automatically proxies `/api/*` requests to the backend.

### Production Build

Build a single binary with embedded frontend:

```bash
make build
```

This creates `bin/claude-proxy` which serves both the frontend and API on port 4000.

Run the binary:
```bash
./bin/claude-proxy
# or with custom config
./bin/claude-proxy server --config config.yaml
```

## Configuration

Copy the example config and customize:

```bash
cp config.example.yaml config.yaml
```

You can also override config with environment variables (uppercase with double underscores):

```bash
export SERVER__PORT=8080
export LOGGER__LEVEL=debug
```

## Development Commands

### Backend

```bash
# Run backend server
make be
go run .

# Run with custom config
go run . server --config custom.yaml

# Run tests
make test
go test ./... -v

# Format code
make format

# Check formatting
make format-check
```

### Frontend

```bash
cd frontend

# Development server
pnpm dev

# Build for production
pnpm build

# Lint
pnpm lint
pnpm lint:fix

# Format
pnpm format
pnpm format:check
```

### Full Stack

```bash
# Build production binary with embedded frontend
make build

# Format all code (Go)
make format

# Run tests (Go)
make test
make test-coverage
```

## Project Structure

```
claude-proxy/
├── cmd/api/              # API server and dependency injection
├── cli/                  # CLI commands
├── config/               # Configuration management
├── pkg/                  # Shared packages
│   ├── errors/          # Custom error handling
│   ├── telegram/        # Telegram integration
│   └── address/         # Address utilities
├── frontend/             # React TypeScript frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── hooks/       # React Query hooks
│   │   ├── lib/         # API client, utilities
│   │   ├── pages/       # Page components
│   │   └── types/       # TypeScript types
│   └── dist/            # Build output (embedded into Go binary)
├── config.yaml          # Runtime configuration
└── main.go              # Application entry point
```

## API Routes

All backend API routes are prefixed with `/api`:

- `GET /api/health` - Health check endpoint

Frontend routes are handled by React Router for client-side navigation.

## Admin Dashboard

The frontend provides an admin interface at `http://localhost:5173` (dev) or `http://localhost:4000` (production):

- **Login** (`/login`) - Authentication
- **Dashboard** (`/admin/dashboard`) - Overview
- **Tokens** (`/admin/tokens`) - API token management
- **App Tokens** (`/admin/app-tokens`) - OAuth application management

## Architecture

### Frontend-Backend Integration

**Development Mode:**
- Frontend: Vite dev server on port 5173
- Backend: Go server on port 4000
- Vite proxies `/api/*` to backend
- Hot reload enabled

**Production Mode:**
- Single binary serves everything on port 4000
- Frontend built and embedded via Go `embed.FS`
- `/api/*` → Backend API handlers
- `/*` → Frontend static files with SPA routing

### Error Handling

The backend uses a panic-based error flow with custom error types:

```go
// In handlers
panic(errors.NewBadRequestError("invalid input", nil))

// Recovery middleware converts to JSON response
// { "error": "invalid input", "code": "BAD_REQUEST" }
```


## Testing

```bash
# Run all tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# Coverage report
make test-coverage  # creates coverage.html
```

## Docker Support

```bash
# Start services
make docker-up

# Stop services
make docker-down

# Run with docker
make dev  # starts docker-compose and runs app
```

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]
