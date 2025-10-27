# Claude Proxy - Multi-Account Claude API Reverse Proxy

A production-ready Claude API reverse proxy with **OAuth 2.0 authentication**, **multi-account support**, **automatic token refresh**, and **load balancing**.

## Features

- **OAuth 2.0 with PKCE**: Secure, scalable OAuth authentication with automatic token refresh
- **Multi-Account Support**: Manage and load-balance across multiple Claude accounts
- **Enhanced Account Status System**: 4-state account management (active, inactive, rate_limited, invalid)
  - Automatic rate limit detection and recovery
  - Invalid token detection with smart error handling
  - Intelligent load balancing that prioritizes healthy accounts
- **Automatic Token Refresh**: Dual triggers - hourly cronjob + on-demand (60-second buffer)
- **Smart Load Balancing**: Stateless round-robin with health filtering and automatic failover
- **Claude API Proxy**: Full proxy support for Claude API requests with SSE streaming
- **Real-time Streaming**: Server-Sent Events (SSE) support for streaming responses
- **Configurable Timeouts**: 5-minute default timeout for extended thinking and long responses
- **Admin Dashboard**: React-based UI with dark/light theme support for OAuth setup and account management
- **Graceful Request Handling**: Smart context cancellation handling - no panics on user-canceled requests
- **JSON Persistence**: File-based account storage (no database required)
- **API Key Protection**: Secure all proxy requests with configurable API keys

## Quick Start

### 1. Prerequisites

- **Go 1.24** (or higher) - [Download Go](https://golang.org/dl/)
  - Uses Go 1.24 language features for improved performance and concurrency
  - Requires Go modules support
- **Claude OAuth Client ID** (obtain from Anthropic)
- **Port 4000** available (configurable)

### 2. Setup

```bash
# Clone and setup
git clone https://github.com/yourusername/claude-proxy.git
cd claude-proxy

# Copy configuration template
cp config.example.yaml config.yaml

# Edit config.yaml
# Required fields:
#   - oauth.client_id: Your Claude OAuth client ID
#   - oauth.token_url: OAuth token endpoint
#   - server.port: Server port (default: 4000)
```

### 3. Build & Run

```bash
# Development (with hot reload for frontend)
# Terminal 1: Backend
go run . server

# Terminal 2: Frontend
cd frontend && pnpm dev

# Production build
go build -o bin/claude-proxy
./bin/claude-proxy
```

Server runs on `http://localhost:4000`

### 4. Add Claude Accounts via Admin Dashboard

**Step 1: Access Admin Dashboard**

```bash
# Open admin UI in browser
http://localhost:4000
```

**Step 2: Authenticate with Admin API Key**

Enter your configured admin API key (from `config.yaml` `auth.api_key` field):
```
API Key: your-configured-api-key
```

**Step 3: Add New Account**

1. Click "Add Account" button
2. Click "Authorize with Claude"
3. Browser opens Claude OAuth authorization page
4. Authorize and approve access
5. Claude redirects back with account tokens
6. Account is automatically saved and ready to use

✅ Account is now saved! Tokens auto-refresh every hour + on-demand (60s before expiry).

**Admin Dashboard Features:**
- View all saved accounts
- See token expiration status
- Monitor account health (active/inactive)
- View last refresh error (if any)
- Manual account management
- Add multiple accounts and switch between them

### 5. Use the Proxy to Send Requests

Once accounts are added via the admin dashboard, clients can send requests through Claude Proxy using standard Claude API format:

```bash
# Send request with Bearer token
curl -X POST http://localhost:4000/v1/messages \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello Claude!"}
    ],
    "stream": true
  }'
```

**How it works:**
1. Request arrives at `/v1/messages` with Bearer token
2. Proxy validates the token
3. Proxy automatically selects a healthy account via load balancing
4. Checks if account token needs refresh (60-second buffer)
5. Refreshes token if needed (transparent to client)
6. Forwards request to Claude API with selected account's token
7. Returns response to client (streaming or JSON)

**Note:** The endpoint `/v1/messages` is compatible with the standard Claude API format, so you can drop in Claude Proxy as a replacement for `https://api.claude.ai`. Use your stored token with `Authorization: Bearer` header.

## ✨ Streaming Support

**Real-time SSE Streaming**: The proxy now supports Server-Sent Events (SSE) for real-time response streaming!

**How it works:**
- Automatically detects `"stream": true` in requests
- Streams Claude API responses in real-time using SSE
- Provides immediate feedback for extended thinking and long responses
- Low memory footprint - no buffering required
- Graceful handling of client disconnections

**Usage:**
```bash
curl -X POST http://localhost:4000/v1/messages \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

## 🛡️ Enhanced Account Status System

The proxy features an intelligent 4-state account management system with automatic error detection and recovery:

### Account States

1. **`active`** - Healthy and available for proxying requests
2. **`inactive`** - Manually disabled by admin
3. **`rate_limited`** - Temporarily unavailable due to Claude API rate limits
4. **`invalid`** - Authentication credentials revoked or expired

### Automatic Error Detection

**Rate Limit Detection (429 Errors)**:
- Automatically detects when Claude API returns 429 status
- Marks account as `rate_limited` with 1-hour recovery period
- Account excluded from load balancing until rate limit expires
- Hourly scheduler automatically recovers expired rate-limited accounts

**Invalid Token Detection (401/403 Errors)**:
- Detects authentication errors from token refresh failures
- Marks account as `invalid` (requires manual intervention)
- Permanently excluded from load balancing until reactivated

### Smart Load Balancing

The proxy intelligently selects accounts based on health status:

1. **Priority 1**: Healthy `active` accounts (not needing token refresh)
2. **Priority 2**: `active` accounts that need token refresh
3. **Priority 3**: Recently recovered `rate_limited` accounts
4. **Excluded**: Current `rate_limited`, `invalid`, and `inactive` accounts

**Error Messages**:
- Clear differentiation: "no accounts available" vs "all accounts rate limited/invalid"
- Detailed logging with account status for debugging

### Automatic Recovery

The scheduler runs hourly to:
1. Check all `rate_limited` accounts for expired limits
2. Automatically recover and mark as `active`
3. Resume routing requests to recovered accounts
4. Log recovery events for monitoring

### API Response Example

Account status is visible in the admin API:

```bash
GET /api/accounts
```

```json
{
  "accounts": [
    {
      "id": "app_123",
      "name": "Production Account",
      "status": "active",
      "rate_limited_until": null,
      "last_refresh_error": "",
      "expires_at": 1735347600
    },
    {
      "id": "app_456",
      "name": "Backup Account",
      "status": "rate_limited",
      "rate_limited_until": 1735351200,
      "last_refresh_error": "failed to refresh token: status 429: rate limit exceeded",
      "expires_at": 1735344000
    }
  ]
}
```

### Benefits

- **Zero Downtime**: Automatic failover when accounts hit rate limits
- **Self-Healing**: Accounts automatically recover without manual intervention
- **Transparent**: Clear status visibility in admin dashboard and API
- **Reliable**: Smart load balancing prevents routing to unhealthy accounts

## API Endpoints

### Admin Authentication

All admin endpoints require the `X-API-Key` header with your configured admin API key:

```bash
-H "X-API-Key: your-configured-api-key"
```

### OAuth Flow (Admin Only)

Used by admin dashboard to add accounts:

- **`GET /oauth/authorize`** - Generate OAuth authorization URL with PKCE
  - Returns: `{ authorization_url, state, code_verifier }`
  - Requires: `X-API-Key` header (admin API key)

- **`POST /oauth/exchange`** - Exchange authorization code for access token
  - Body: `{ "code": "...", "state": "...", "code_verifier": "..." }`
  - Returns: Account info with tokens and expiry
  - Saves account to JSON persistence
  - Requires: `X-API-Key` header (admin API key)

### Account Management

- **`GET /api/accounts`** - List all saved accounts
  - Requires: `X-API-Key` header
  - Returns: Array of accounts with status, tokens, and expiry info
  - Shows: account health (`active`, `inactive`, `rate_limited`, `invalid`)
  - Includes: `rate_limited_until` timestamp and `last_refresh_error` message

- **`POST /api/accounts`** - Create new account from OAuth exchange
  - Requires: `X-API-Key` header
  - Body: `{ "code": "...", "state": "...", "code_verifier": "..." }`
  - Returns: New account with tokens

- **`PUT /api/accounts/{id}`** - Update account status or name
  - Requires: `X-API-Key` header
  - Body: `{ "name": "...", "status": "active|inactive|rate_limited|invalid" }`
  - Allows manual status changes for recovery or maintenance

- **`DELETE /api/accounts/{id}`** - Remove account
  - Requires: `X-API-Key` header
  - Stops routing requests to this account

### Claude API Proxy

- **`POST /v1/messages`** (and all `/v1/*` endpoints) - Proxy Claude API requests
  - Requires: `Authorization: Bearer <token>` header
  - Body: Standard Claude API request format
    - `model`: Model identifier (e.g., "claude-opus-4-20250514")
    - `messages`: Array of messages with role and content
    - `max_tokens`: Maximum tokens in response
    - `stream`: Boolean for streaming response (optional)
  - Auto-selects healthy account via load balancing
  - Auto-refreshes token if within 60 seconds of expiry
  - Returns: Claude API response in standard format (streaming or JSON)

### Health & Status

- **`GET /health`** - Health check endpoint
  - No auth required
  - Returns: Server status and account health summary

## Configuration

See `config.example.yaml` for template. Key sections:

```yaml
server:
  host: "0.0.0.0"
  port: 4000
  # Request timeout for LLM API requests (default: 5m)
  # Recommended: 5m for extended thinking, 10m for very long tasks
  request_timeout: 5m

oauth:
  client_id: "your-claude-oauth-client-id"
  authorize_url: "https://claude.ai/oauth/authorize"
  token_url: "https://api.claude.ai/oauth/token"
  redirect_uri: "http://localhost:4000/oauth/callback"
  scope: "user:profile user:inference"

claude:
  base_url: "https://api.claude.ai"

storage:
  data_folder: "~/.claude-proxy/data"

auth:
  api_key: "your-secret-api-key"

logger:
  level: "info"
  format: "text"

retry:
  max_retries: 3
  retry_delay: "1s"
```

## Environment Variables

Override any YAML config with uppercase env vars using `__` for nesting:

```bash
# Server
export SERVER__PORT=8080
export SERVER__HOST=127.0.0.1
export SERVER__REQUEST_TIMEOUT=10m  # For very long requests (default: 5m)

# OAuth
export OAUTH__CLIENT_ID=your-client-id
export OAUTH__TOKEN_URL=https://api.claude.ai/oauth/token

# Auth
export AUTH__API_KEY=your-secret-key

# Storage
export STORAGE__DATA_FOLDER=~/.claude-proxy/data

# Logger
export LOGGER__LEVEL=debug
export LOGGER__FORMAT=json
```

## Data Storage

Account credentials stored in `~/.claude-proxy/data/` as JSON:

```
~/.claude-proxy/data/
├── account_*.json          # Individual account files
└── ...
```

Each account contains:
- `id`: Unique account identifier
- `name`: Account display name
- `access_token`: Current access token
- `refresh_token`: For obtaining new access tokens
- `expires_at`: Timestamp when access token expires
- `refresh_at`: Timestamp of last token refresh
- `status`: active/inactive
- `last_refresh_error`: Error message if refresh failed

**⚠️ SECURITY**: Keep `~/.claude-proxy/data/` secure (0700 permissions). Files contain sensitive OAuth tokens.

## Admin Dashboard

The admin dashboard is a modern React application with the following features:

**UI Features:**
- **Dark/Light Theme**: Automatic theme switching with system preference detection and manual override
- **Responsive Design**: Mobile-friendly interface using TailwindCSS v4
- **Real-time Updates**: React Query for efficient data fetching and caching
- **Account Management**: View all accounts, their status, and token expiration
- **OAuth Flow**: Guided OAuth setup process with visual feedback
- **Token Management**: Create, edit, and delete API tokens with usage tracking

**Theme Support:**
- Three modes: Light, Dark, and System (follows OS preference)
- Persistent theme selection (stored in localStorage)
- Optimized color contrast for readability in both modes
- Smooth theme transitions

**Tech Stack:**
- React 19 with TypeScript
- Vite 7 for fast builds and HMR
- TailwindCSS v4 with shadcn/ui components
- React Router DOM v7 for routing
- TanStack React Query v5 for state management
- ESLint v9 with Prettier integration

## Development

**Backend:**
```bash
go run . server
# or
make be
```

**Frontend:**
```bash
cd frontend && pnpm dev
# or
make fe
```

**Full Stack** (two terminals):
```bash
# Terminal 1
make be

# Terminal 2
make fe
```

**Build Production Binary:**
```bash
cd frontend && pnpm build  # Build React
go build -o bin/claude-proxy  # Embed frontend + build Go
# or
make build
```

**Testing & Formatting:**
```bash
# Backend
go test ./...           # Run all tests
go test ./modules/...   # Test specific module
make format             # Format Go code
make test-coverage      # Generate coverage report

# Frontend
cd frontend && pnpm lint        # Lint check
cd frontend && pnpm lint:fix    # Lint and fix
cd frontend && pnpm format      # Format code
cd frontend && pnpm build       # Production build
```

## Architecture

**Request Flow:**
```
┌──────────────────────┐
│   Client Request     │
│   (X-API-Key Auth)   │
└──────────┬───────────┘
           ▼
┌──────────────────────────────────┐
│   Claude Proxy Server            │
│  ┌─────────────────────────────┐ │
│  │ Multi-Account Load Balancer  │ │
│  │ (Round-Robin + Health Check) │ │
│  └────────────┬────────────────┘ │
└───────────────┼──────────────────┘
                │ Selected Account
                ▼
    ┌──────────────────────────┐
    │ Token Refresh Check      │
    │ (60s before expiry)      │
    └────────────┬─────────────┘
                 │
        ┌────────┴────────┐
        ▼                 ▼
   [Need Refresh]    [Token Valid]
        │                 │
        ▼                 │
   ┌─────────┐           │
   │  OAuth  │           │
   │ Refresh │           │
   └────┬────┘           │
        │                │
        └────────┬───────┘
                 ▼
      ┌────────────────────┐
      │  Claude API        │
      │  (Proxy Request)   │
      └────────────────────┘
```

**Key Components:**
- **Scheduler**: Cronjob runs hourly to refresh expiring tokens
- **Load Balancer**: Stateless round-robin account selection
- **Token Refresh**: Automatic on-demand + scheduled
- **OAuth Service**: PKCE-based token exchange and refresh
- **JSON Persistence**: File-based account storage (no database)
- **Admin Dashboard**: React UI for account/OAuth management

## Go 1.24 Features

This project leverages modern Go 1.24 features for improved performance, concurrency, and developer experience:

- **Enhanced Concurrency**: Uses Go 1.24's optimized goroutine scheduling and improved `sync` primitives
- **Better Error Handling**: Leverages Go's error wrapping and chain capabilities
- **Generic Collections**: Uses Go generics for type-safe data handling in repositories and services
- **Range Over Integer**: Modern syntax for simple numeric loops (e.g., `for i := range n`)
- **Improved Standard Library**: Benefits from latest Go standard library improvements and optimizations
- **Stronger Memory Safety**: Compiler improvements for better memory safety checks
- **Better Performance**: Go 1.24 runtime improvements for faster request handling and token refresh operations

**Version Requirements:**
```bash
# Check Go version
go version

# Must be Go 1.24 or higher
# Recommended: Go 1.24.0 or latest stable
```

If you have an older Go version, [download Go 1.24+](https://golang.org/dl/).

## Error Handling

**Graceful Request Cancellation:**
- Smart detection of user-canceled requests (context cancellation)
- No panic recovery logs for normal user cancellations
- Proper HTTP status codes:
  - `499` - Client Closed Request (user canceled)
  - `408` - Request Timeout (deadline exceeded)
  - `503` - Service Unavailable (actual errors)

**Robust Error Recovery:**
- Panic recovery middleware catches unexpected errors
- AppError system for structured, HTTP-aware error handling
- Automatic retry logic with exponential backoff
- Detailed error logging for debugging

## Security

- **API Key Authentication**: All proxy requests require valid API key
- **OAuth 2.0 + PKCE**: Secure, standards-compliant authentication
- **Automatic Token Refresh**: 60-second buffer prevents token expiry
- **File Permissions**: Account data stored with restricted 0700 permissions
- **No Token Exposure**: Tokens never logged or exposed in responses
- **HTTPS Ready**: Configure with reverse proxy for HTTPS in production
- **Context-Aware Request Handling**: Graceful handling of canceled and timed-out requests

## Token Refresh

- **Automatic Hourly**: Cronjob runs at 0 minutes every hour
- **On-Demand**: Triggers when token within 60 seconds of expiry
- **Transparent**: Refresh happens automatically, no user intervention needed
- **Error Handling**: Failed refreshes logged, account marked as unhealthy
- **Load Balancer Aware**: Prefers accounts with valid tokens for better UX

## License

MIT

## Contributing

Contributions welcome! Please:
1. Follow existing code patterns (DDD, dependency injection)
2. Add tests for new features
3. Update documentation
4. Test both backend and frontend changes

## Support

For issues and questions:
- **GitHub Issues**: [Report Issue](https://github.com/yourusername/claude-proxy/issues)
- **Documentation**: See `CLAUDE.md` for architecture and `docs/` for guides

