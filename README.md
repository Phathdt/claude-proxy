# Claude Proxy - Multi-Account Claude API Reverse Proxy

A production-ready Claude API reverse proxy with **OAuth 2.0 authentication**, **multi-account support**, **automatic token refresh**, and **load balancing**.

## Features

- **OAuth 2.0 with PKCE**: Secure, scalable OAuth authentication with automatic token refresh
- **Multi-Account Support**: Manage and load-balance across multiple Claude accounts
- **Enhanced Account Status System**: 4-state account management (active, inactive, rate_limited, invalid)
  - Automatic rate limit detection and recovery
  - Invalid token detection with smart error handling
  - Intelligent load balancing that prioritizes healthy accounts
- **Session Limiting**: Prevent abuse with configurable concurrent session limits per client (IP + UserAgent)
  - JSON file-based session tracking (no Redis required)
  - Automatic session expiry and cleanup
  - Admin dashboard for session monitoring
  - Dynamic account rotation per request
- **Automatic Token Refresh**: Dual triggers - hourly cronjob + on-demand (60-second buffer)
- **Smart Load Balancing**: Stateless round-robin with health filtering and automatic failover
- **Claude API Proxy**: Full proxy support for Claude API requests with SSE streaming
- **Real-time Streaming**: Server-Sent Events (SSE) support for streaming responses
- **Configurable Timeouts**: 5-minute default timeout for extended thinking and long responses
- **Admin Dashboard**: React-based UI with dark/light theme support for OAuth setup and account management
- **Graceful Request Handling**: Smart context cancellation handling - no panics on user-canceled requests
- **JSON Persistence**: File-based account and session storage (no database required)
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
git clone https://github.com/phathdt379/claude-proxy.git
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
- Monitor account health (active/inactive/rate_limited/invalid)
- Real-time statistics dashboard with system health monitoring
- View last refresh error (if any)
- Manual account management
- Add multiple accounts and switch between them

### 5. Use the Proxy

Send requests to `/v1/messages` with `Authorization: Bearer <token>` header. The proxy automatically:

- Selects a healthy account via load balancing
- Refreshes tokens if needed (60-second buffer)
- Forwards requests to Claude API
- Returns streaming or JSON responses

Compatible with standard Claude API format - drop-in replacement for `https://api.claude.ai`.

## ✨ Streaming Support

Supports Server-Sent Events (SSE) for real-time response streaming. Set `"stream": true` in requests for immediate feedback on extended thinking and long responses.

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

The scheduler runs hourly to check `rate_limited` accounts, automatically recover expired limits, and resume routing requests.

### Benefits

- **Zero Downtime**: Automatic failover when accounts hit rate limits
- **Self-Healing**: Accounts automatically recover without manual intervention
- **Transparent**: Clear status visibility in admin dashboard and API
- **Reliable**: Smart load balancing prevents routing to unhealthy accounts

## API Endpoints

All admin endpoints require `X-API-Key` header with your configured admin API key.

### Account Management

- **`GET /api/accounts`** - List all accounts with status and token info
- **`POST /api/accounts`** - Create new account from OAuth exchange
- **`PUT /api/accounts/{id}`** - Update account status or name
- **`DELETE /api/accounts/{id}`** - Remove account

### Claude API Proxy

- **`POST /v1/messages`** (and all `/v1/*`) - Proxy requests to Claude API
  - Requires: `Authorization: Bearer <token>` header
  - Auto-selects healthy account, auto-refreshes tokens, returns streaming or JSON

### Admin & Monitoring

- **`GET /api/admin/statistics`** - System statistics and health metrics
  - Returns: Account counts by status, token health, system health (`healthy`/`degraded`/`unhealthy`)

### Health Check

- **`GET /health`** - Server status (no auth required)

## Environment Variables

Override YAML config with uppercase env vars using `__` for nesting:

```bash
AUTH__API_KEY=your-secret-key
LOGGER__LEVEL=debug
```

## Data Storage

Account credentials stored in `~/.claude-proxy/data/` as JSON files.

**⚠️ SECURITY**: Keep `~/.claude-proxy/data/` secure (0700 permissions). Contains sensitive OAuth tokens.

## Admin Dashboard

Modern React application with:

- Dark/Light theme support
- Real-time statistics with 30s auto-refresh
- Account management with OAuth flow
- System health monitoring
- Responsive design with TailwindCSS v4

**Tech Stack**: React 19 + TypeScript, Vite 7, TanStack Query v5, shadcn/ui

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
```

**Build Production Binary:**

```bash
make build
```

## Architecture

**Key Components:**

- **Load Balancer**: Round-robin with health-aware account selection
- **Token Refresh**: Hourly scheduler + on-demand (60s buffer)
- **OAuth Service**: PKCE-based token exchange and refresh
- **JSON Persistence**: File-based storage (no database)
- **Admin Dashboard**: React UI for management

**Request Flow**: Client → Load Balancer → Token Check → OAuth Refresh (if needed) → Claude API

## Roadmap

**Feature Parity**: 10/12 (83% complete) with Python version

**Next Up**: Idle Account Detection, Enhanced Exponential Backoff

See [ROADMAP.md](ROADMAP.md) for detailed feature comparison, implementation plans, and future enhancements.

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

- **GitHub Issues**: [Report Issue](https://github.com/phathdt379/claude-proxy/issues)
- **Documentation**: See `CLAUDE.md` for architecture and `docs/` for guides
