# Claude Proxy - Multi-Account Claude API Reverse Proxy

A production-ready Claude API reverse proxy with **OAuth 2.0 authentication**, **multi-account support**, **automatic token refresh**, and **load balancing**.

## Features

- **OAuth 2.0 with PKCE**: Secure, scalable OAuth authentication with automatic token refresh
- **Multi-Account Support**: Manage and load-balance across multiple Claude accounts
- **Automatic Token Refresh**: Dual triggers - hourly cronjob + on-demand (60-second buffer)
- **Load Balancing**: Stateless round-robin account selection with health filtering
- **Claude API Proxy**: Full proxy support for Claude API requests with streaming
- **Admin Dashboard**: React-based UI for OAuth setup and account management
- **JSON Persistence**: File-based account storage (no database required)
- **API Key Protection**: Secure all proxy requests with configurable API keys

## Quick Start

### 1. Prerequisites

- **Go 1.21+**
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
# Send request with your API key (just like Claude API)
curl -X POST http://localhost:4000/v1/messages \
  -H "X-API-Key: your-api-key" \
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
1. Request arrives at `/v1/messages` with your API key
2. Proxy automatically selects a healthy account via load balancing
3. Checks if account token needs refresh (60-second buffer)
4. Refreshes token if needed (transparent to client)
5. Forwards request to Claude API with selected account's token
6. Returns response to client (streaming or JSON)

**Note:** The endpoint `/v1/messages` is compatible with the standard Claude API format, so you can drop in Claude Proxy as a replacement for `https://api.claude.ai`.

## API Endpoints

### Admin Authentication

All admin endpoints require the `X-API-Key` header with your configured admin API key:

```bash
-H "X-API-Key: your-configured-api-key"
```

### OAuth Flow (Internal)

Used by admin dashboard to add accounts:

- **`GET /oauth/authorize`** - Generate OAuth authorization URL with PKCE
  - Returns: `{ authorization_url, state, code_verifier }`
  - Requires: `X-API-Key` header

- **`POST /oauth/exchange`** - Exchange authorization code for access token
  - Body: `{ "code": "...", "state": "...", "code_verifier": "..." }`
  - Returns: Account info with tokens and expiry
  - Saves account to JSON persistence
  - Requires: `X-API-Key` header

- **`GET /oauth/callback`** - OAuth callback handler
  - Receives: `?code=AUTH_CODE&state=STATE`

### Account Management

- **`GET /api/accounts`** - List all saved accounts
  - Requires: `X-API-Key` header
  - Returns: Array of accounts with status, tokens, and expiry info
  - Shows: account health, last refresh time, and any errors

- **`POST /api/accounts`** - Create new account from OAuth exchange
  - Requires: `X-API-Key` header
  - Body: `{ "code": "...", "state": "...", "code_verifier": "..." }`
  - Returns: New account with tokens

- **`DELETE /api/accounts/{id}`** - Remove account
  - Requires: `X-API-Key` header
  - Stops routing requests to this account

### Proxy Requests

- **`POST /v1/messages`** - Proxy requests to Claude API (standard Claude API format)
  - Requires: `X-API-Key` header (use your configured API key, not OAuth token)
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
go test ./...           # Run all tests
go test ./modules/...   # Test specific module
make format             # Format Go code
make test-coverage      # Generate coverage report

cd frontend && pnpm lint:fix  # Lint/fix frontend
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

## Security

- **API Key Authentication**: All proxy requests require valid API key
- **OAuth 2.0 + PKCE**: Secure, standards-compliant authentication
- **Automatic Token Refresh**: 60-second buffer prevents token expiry
- **File Permissions**: Account data stored with restricted 0700 permissions
- **No Token Exposure**: Tokens never logged or exposed in responses
- **HTTPS Ready**: Configure with reverse proxy for HTTPS in production

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

