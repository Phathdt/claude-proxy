# Clove - Claude API Reverse Proxy

A lightweight, production-ready Claude reverse proxy with OAuth 2.0 authentication and automatic token refresh.

## Features

- **OAuth 2.0 with PKCE**: Secure authentication flow with automatic token refresh
- **OpenAI-Compatible API**: Drop-in replacement with `/v1/messages` endpoint
- **Streaming Support**: Real-time Server-Sent Events (SSE) streaming
- **Automatic Token Refresh**: Tokens refreshed automatically before expiry
- **Single Account Management**: JSON file-based persistence
- **API Key Protection**: Secure your proxy with API key authentication

## Quick Start

### 1. Prerequisites

- Go 1.24+ 
- Claude OAuth Client ID (contact Anthropic for API access)

### 2. Configuration

```bash
# Copy example configuration
cp config.example.yaml config.yaml

# Edit config.yaml with your settings
# Required:
#   - oauth.client_id: Your Claude OAuth client ID
#   - auth.api_key: Your chosen API key for protecting the proxy
```

### 3. Run the Server

```bash
# Build and run
go build -o clove
./clove

# Or run directly
go run . server
```

The server will start on `http://localhost:5201`

### 4. Setup OAuth Authentication

1. Get authorization URL:
```bash
curl http://localhost:5201/oauth/authorize
```

2. Visit the returned `authorization_url` in your browser and authorize

3. After authorization, you'll be redirected to the callback URL and your account will be saved

### 5. Send Messages

```bash
curl -X POST http://localhost:5201/v1/messages \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-20250514",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "max_tokens": 1024
  }'
```

## API Endpoints

### OAuth Flow

- `GET /oauth/authorize` - Get OAuth authorization URL
- `GET /oauth/callback` - OAuth callback handler (automatic)

### Claude API (Requires `X-API-Key` header)

- `POST /v1/messages` - Send message to Claude
  - Supports `stream: true` for SSE streaming
  - Supports `stream: false` for JSON response

### Health Check

- `GET /health` - Health check with account status

## Configuration

See `config.example.yaml` for all available options:

- **Server**: Host and port configuration
- **OAuth**: Claude OAuth 2.0 settings
- **Auth**: API key for proxy authentication
- **Claude**: Claude API base URL
- **Storage**: Data folder for account persistence
- **Retry**: Retry configuration for failed requests
- **Logger**: Logging level and format

## Environment Variables

Override config with environment variables using double underscore:

```bash
export SERVER__PORT=8080
export AUTH__API_KEY=my-secret-key
export OAUTH__CLIENT_ID=your-client-id
```

## Data Storage

Account data is stored in `~/.clove/data/account.json` with:
- Access token
- Refresh token
- Organization UUID
- Token expiry timestamp

**⚠️ Keep this file secure - it contains sensitive credentials**

## Development

```bash
# Install dependencies
go mod download

# Run with custom config
go run . server --config custom.yaml

# Build
go build -o clove

# Format code
go fmt ./...

# Run tests
go test ./...
```

## Architecture

```
┌─────────────────┐
│   Client App    │
└────────┬────────┘
         │ X-API-Key
         ▼
┌─────────────────┐
│  Clove Proxy    │
│  - OAuth PKCE   │
│  - Auto Refresh │
│  - SSE Stream   │
└────────┬────────┘
         │ Bearer Token
         ▼
┌─────────────────┐
│  Claude API     │
└─────────────────┘
```

## Security

- API key authentication for all proxy requests
- OAuth 2.0 with PKCE for Claude authentication
- Automatic token refresh with 60-second buffer
- Account data stored with 0600 permissions
- No token logging or exposure

## License

MIT

## Support

For issues and questions:
- GitHub Issues: [Report Issue](https://github.com/yourusername/clove/issues)
- Documentation: See `docs/mvp.md` for detailed specifications

