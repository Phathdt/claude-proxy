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

### 4. Add Claude Accounts via OAuth

**Option A: Via Admin Dashboard (Recommended)**

1. Open `http://localhost:4000` in browser
2. Login (username/password can be anything initially)
3. Click "Add Account" and follow OAuth flow
4. Authorize with Claude
5. Account tokens are automatically saved

**Option B: Via API (Manual Flow)**

```bash
# Step 1: Get OAuth authorization URL
curl http://localhost:4000/oauth/authorize

# Response includes:
# - authorization_url: Visit this in browser to authorize
# - state: Save this
# - code_verifier: Save this

# Step 2: Visit authorization_url and authorize
# (Browser redirects to http://localhost:4000/oauth/callback?code=AUTH_CODE&state=...)

# Step 3: Exchange code for tokens
curl -X POST http://localhost:4000/oauth/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "code": "AUTH_CODE_FROM_CALLBACK",
    "state": "STATE_FROM_STEP_1",
    "code_verifier": "CODE_VERIFIER_FROM_STEP_1"
  }'
```

✅ Account is now saved! Tokens auto-refresh every hour + on-demand (60s before expiry).

### 5. Send Requests to Claude

```bash
# Using API key for authentication
curl -X POST http://localhost:4000/api/proxy \
  -H "X-API-Key: your-configured-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "Hello Claude!"}
    ],
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "stream": true
  }'
```

The request is automatically routed to an available Claude account with a valid token.

## API Endpoints

### OAuth Authentication

- **`GET /oauth/authorize`** - Generate OAuth authorization URL with PKCE
  - Returns: `{ authorization_url, state, code_verifier }`
  - No auth required

- **`POST /oauth/exchange`** - Exchange authorization code for access token
  - Body: `{ "code": "...", "state": "...", "code_verifier": "..." }`
  - Returns: Account info with tokens and expiry
  - Saves account to JSON persistence
  - No auth required

- **`GET /oauth/callback`** - OAuth callback handler
  - Receives: `?code=AUTH_CODE&state=STATE`

### Account Management

- **`GET /api/accounts`** - List all saved accounts
  - Requires: `X-API-Key` header
  - Returns: Array of accounts with status and expiry info

- **`POST /api/accounts`** - Create account from OAuth exchange
  - Requires: `X-API-Key` header
  - Body: OAuth exchange payload

### Proxy Requests

- **`GET /api/proxy/*`** - Proxy requests to Claude API
  - Requires: `X-API-Key` header
  - Automatically selects healthy account via load balancing
  - Refreshes token if within 60 seconds of expiry
  - Returns: Claude API response (streaming or JSON)

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

