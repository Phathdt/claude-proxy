# Clove MVP - OAuth Edition

> Minimum Viable Product specification for Clove with OAuth authentication and refresh token support

## Table of Contents

- [MVP Overview](#mvp-overview)
- [Scope Definition](#scope-definition)
- [Core Features](#core-features)
- [Architecture](#architecture)
- [Implementation Phases](#implementation-phases)
- [API Endpoints](#api-endpoints)
- [File Structure](#file-structure)
- [Data Models](#data-models)
- [Configuration](#configuration)
- [Authentication Flow](#authentication-flow)
- [Error Handling](#error-handling)
- [Testing Strategy](#testing-strategy)
- [Future Enhancements](#future-enhancements)

---

## MVP Overview

### Goal

Create a lightweight, production-ready Claude reverse proxy that uses OAuth 2.0 authentication with PKCE and refresh token support to access the Claude API.

### Target Users

- Developers who want to use Claude API through a standardized interface
- Applications requiring Claude integration with OAuth security
- Teams needing a self-hosted Claude proxy solution

### Success Criteria

- [ ] Successfully authenticate via OAuth 2.0 with PKCE
- [ ] Automatic token refresh when access token expires
- [ ] Send messages to Claude API (streaming & non-streaming)
- [ ] Handle basic errors gracefully
- [ ] Simple configuration via environment variables
- [ ] Single account support working reliably
- [ ] Response time < 2s for simple messages

---

## Scope Definition

### ✅ Included in MVP

#### 1. OAuth Authentication

- OAuth 2.0 with PKCE flow
- Access token management
- **Refresh token handling and automatic renewal**
- Token expiry detection
- Organization UUID extraction

#### 2. Basic API

- `POST /v1/messages` endpoint
- OpenAI-compatible request format
- Streaming responses (SSE)
- Non-streaming responses (JSON)
- Basic request validation

#### 3. Account Management

- Single account storage (JSON file)
- Account persistence
- Basic account status (valid/invalid)

#### 4. Message Processing

- Text content support
- Simple message formatting
- Request/response handling

#### 5. Configuration

- Environment variable support
- JSON config file (`~/.clove/data/config.json`)
- Basic settings (host, port, API key)

#### 6. Error Handling

- HTTP error responses
- OAuth errors
- API errors
- Basic retry logic (3 attempts)

#### 7. Health Check

- `GET /health` endpoint
- Basic status reporting

### ❌ Excluded from MVP (Future Enhancements)

- Multi-account load balancing
- Web proxy mode (web.claude.ai fallback)
- Cookie-based authentication
- Tool/function calling
- Prompt caching
- Extended thinking support
- Web search integration
- Stop sequences
- Token counting
- Admin web interface
- Multi-language support (i18n)
- Session management complexity
- SillyTavern compatibility mode
- Image support (text-only for MVP)
- File uploads
- Custom system prompts
- Static file serving
- Multiple HTTP backends (httpx only)
- Rate limiting middleware
- Statistics dashboard
- Advanced logging (basic logging only)

---

## Core Features

### 1. OAuth 2.0 with PKCE and Refresh Token

**File**: `app/services/oauth.py`

```python
class OAuthService:
    def __init__(self):
        self.client_id = config.OAUTH_CLIENT_ID
        self.authorize_url = config.OAUTH_AUTHORIZE_URL
        self.token_url = config.OAUTH_TOKEN_URL
        self.redirect_uri = config.OAUTH_REDIRECT_URI

    async def exchange_code_for_token(self, code: str, code_verifier: str) -> dict:
        """Exchange authorization code for access + refresh tokens"""
        pass

    async def refresh_access_token(self, refresh_token: str) -> dict:
        """Use refresh token to get new access token"""
        pass

    def is_token_expired(self, expires_at: int) -> bool:
        """Check if access token is expired"""
        pass

    async def get_organization_uuid(self, access_token: str) -> str:
        """Fetch organization UUID from Claude API"""
        pass
```

**Features**:

- PKCE (code_verifier, code_challenge)
- Access token storage
- **Refresh token storage**
- **Automatic token refresh before expiry**
- Token expiry checking
- Secure token exchange

### 2. Messages API

**File**: `app/api/routes/messages.py`

```python
@router.post("/v1/messages")
async def create_message(
    request: MessageRequest,
    stream: bool = False,
    x_api_key: str = Header(None)
):
    """
    Create a message using Claude API
    Supports streaming and non-streaming modes
    """
    pass
```

**Features**:

- OpenAI-compatible input format
- Streaming via SSE
- Non-streaming JSON response
- Automatic token refresh if expired

### 3. Simple Account Management

**File**: `app/core/account.py`

```python
class Account:
    organization_uuid: str
    access_token: str
    refresh_token: str
    expires_at: int  # Unix timestamp
    status: str  # "valid" or "invalid"

class AccountManager:
    def load_account(self) -> Account:
        """Load account from JSON file"""
        pass

    def save_account(self, account: Account):
        """Save account to JSON file"""
        pass

    async def get_valid_token(self) -> str:
        """Get valid access token, refresh if needed"""
        pass
```

### 4. Configuration

**File**: `app/core/config.py`

```python
class Settings:
    # Server
    HOST: str = "0.0.0.0"
    PORT: int = 5201

    # Authentication
    API_KEY: str  # For protecting the proxy

    # OAuth
    OAUTH_CLIENT_ID: str
    OAUTH_AUTHORIZE_URL: str = "https://claude.ai/oauth/authorize"
    OAUTH_TOKEN_URL: str = "https://api.claude.ai/oauth/token"
    OAUTH_REDIRECT_URI: str = "http://localhost:5201/oauth/callback"

    # Claude API
    CLAUDE_API_BASEURL: str = "https://api.claude.ai"

    # Storage
    DATA_FOLDER: str = "~/.clove/data"

    # Retry
    MAX_RETRIES: int = 3
    RETRY_DELAY: int = 1
```

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────┐
│                   Client Application                │
└────────────────────┬────────────────────────────────┘
                     │ HTTP Request
                     │ (X-API-Key header)
                     ▼
┌─────────────────────────────────────────────────────┐
│              FastAPI Application                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  API Key Validation Middleware                │  │
│  └───────────────┬───────────────────────────────┘  │
│                  ▼                                   │
│  ┌───────────────────────────────────────────────┐  │
│  │  POST /v1/messages Endpoint                   │  │
│  └───────────────┬───────────────────────────────┘  │
│                  ▼                                   │
│  ┌───────────────────────────────────────────────┐  │
│  │  Account Manager                              │  │
│  │  - Check token expiry                         │  │
│  │  - Refresh if needed                          │  │
│  │  - Return valid access token                  │  │
│  └───────────────┬───────────────────────────────┘  │
│                  ▼                                   │
│  ┌───────────────────────────────────────────────┐  │
│  │  Claude API Client                            │  │
│  │  - Build request with OAuth token             │  │
│  │  - Send to Claude API                         │  │
│  │  - Handle response (stream/non-stream)        │  │
│  └───────────────┬───────────────────────────────┘  │
└──────────────────┼───────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│              Claude API (api.claude.ai)             │
│  - Validates OAuth token                            │
│  - Processes message                                │
│  - Returns response                                 │
└─────────────────────────────────────────────────────┘
```

### Token Refresh Flow

```
┌──────────┐                                    ┌──────────────┐
│  Client  │                                    │ Clove Proxy  │
└────┬─────┘                                    └──────┬───────┘
     │                                                  │
     │ 1. POST /v1/messages                            │
     │─────────────────────────────────────────────────>│
     │                                                  │
     │                                                  │ 2. Check token expiry
     │                                                  │────────┐
     │                                                  │        │
     │                                                  │<───────┘
     │                                                  │
     │                                                  │ 3. Token expired?
     │                                                  │────────┐
     │                                                  │        │
     │                                                  │<───────┘
     │                                                  │
     │                                                  │ 4. POST /oauth/token
     │                                                  │    (refresh_token)
     │                                        ┌─────────┼─────────────────>
     │                                        │         │              Claude API
     │                                        │         │<─────────────────────┐
     │                                        │         │                      │
     │                                        │         │ 5. New access token  │
     │                                        │         │<─────────────────────┘
     │                                        │         │
     │                                        │         │ 6. Save new token
     │                                        │         │────────┐
     │                                        │         │        │
     │                                        │         │<───────┘
     │                                        │         │
     │                                        │         │ 7. POST /messages
     │                                        │         │    (with new token)
     │                                        │         │─────────────────────>
     │                                        │         │              Claude API
     │                                        │         │<─────────────────────┤
     │                                        │         │                      │
     │                                        │         │ 8. Claude response   │
     │                                        └─────────┼───────────<──────────┘
     │                                                  │
     │ 9. Stream/Return response                       │
     │<─────────────────────────────────────────────────│
     │                                                  │
```

---

## Implementation Phases

### Phase 1: Project Setup (Day 1)

**Effort**: 2-4 hours

- [ ] Create project structure
- [ ] Set up FastAPI application
- [ ] Configure dependency management (Poetry/pip)
- [ ] Set up basic logging
- [ ] Create configuration loader
- [ ] Initialize data folder structure

**Deliverable**: Running FastAPI server with health check endpoint

### Phase 2: OAuth Implementation (Day 1-2)

**Effort**: 6-8 hours

- [ ] Implement OAuth service
  - [ ] Code verifier/challenge generation
  - [ ] Authorization URL builder
  - [ ] Token exchange endpoint
  - [ ] **Refresh token exchange**
  - [ ] **Token expiry checking**
- [ ] Create account model
- [ ] Implement account storage (JSON)
- [ ] Add organization UUID fetching
- [ ] Create OAuth callback endpoint

**Deliverable**: Working OAuth flow with token refresh

### Phase 3: Account Management (Day 2)

**Effort**: 3-4 hours

- [ ] Build AccountManager singleton
- [ ] Implement load/save account
- [ ] Add token validation
- [ ] **Implement automatic token refresh logic**
- [ ] Add account status checking

**Deliverable**: Persistent account with auto-refresh

### Phase 4: Messages API (Day 2-3)

**Effort**: 6-8 hours

- [ ] Create message models (Pydantic)
- [ ] Implement Claude API client
  - [ ] Request builder
  - [ ] HTTP client (httpx)
  - [ ] OAuth header injection
- [ ] Build messages endpoint
- [ ] Add request validation
- [ ] Implement retry logic

**Deliverable**: Basic message sending works

### Phase 5: Streaming Support (Day 3)

**Effort**: 4-6 hours

- [ ] Implement SSE streaming
- [ ] Parse Claude streaming events
- [ ] Build streaming response handler
- [ ] Add non-streaming fallback
- [ ] Test both modes

**Deliverable**: Streaming and non-streaming responses

### Phase 6: Error Handling (Day 4)

**Effort**: 3-4 hours

- [ ] Create custom exceptions
- [ ] Add error response formatter
- [ ] Implement retry logic
- [ ] Handle OAuth errors
- [ ] Handle API errors
- [ ] Add proper logging

**Deliverable**: Robust error handling

### Phase 7: Security & Authentication (Day 4)

**Effort**: 2-3 hours

- [ ] Add API key middleware
- [ ] Implement key validation
- [ ] Secure token storage
- [ ] Add CORS configuration

**Deliverable**: Secure proxy

### Phase 8: Testing & Documentation (Day 5)

**Effort**: 4-6 hours

- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Create API documentation
- [ ] Write setup guide
- [ ] Add example usage

**Deliverable**: Tested and documented MVP

**Total Estimated Time**: 5 days (30-40 hours)

---

## API Endpoints

### 1. Health Check

```http
GET /health
```

**Response**:

```json
{
  "status": "ok",
  "account_valid": true
}
```

### 2. Create Message

```http
POST /v1/messages
X-API-Key: your-api-key
Content-Type: application/json
```

**Request**:

```json
{
  "model": "claude-opus-4-20250514",
  "messages": [
    {
      "role": "user",
      "content": "Hello, Claude!"
    }
  ],
  "max_tokens": 1024,
  "stream": false
}
```

**Response (Non-streaming)**:

```json
{
  "id": "msg_123",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "Hello! How can I help you today?"
    }
  ],
  "model": "claude-opus-4-20250514",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 20
  }
}
```

**Response (Streaming)**:

```
event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","content":[],"model":"claude-opus-4-20250514"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":20}}

event: message_stop
data: {"type":"message_stop"}
```

### 3. OAuth Callback

```http
GET /oauth/callback?code=AUTH_CODE&state=STATE
```

**Response**:

```json
{
  "success": true,
  "message": "Account configured successfully",
  "organization_uuid": "org_123..."
}
```

### 4. OAuth Authorization URL

```http
GET /oauth/authorize
```

**Response**:

```json
{
  "authorization_url": "https://claude.ai/oauth/authorize?client_id=...&state=...&code_challenge=..."
}
```

---

## File Structure

```
clove-mvp/
├── app/
│   ├── __init__.py
│   ├── main.py                      # FastAPI app entry point
│   │
│   ├── api/
│   │   ├── __init__.py
│   │   └── routes/
│   │       ├── __init__.py
│   │       ├── messages.py          # POST /v1/messages
│   │       ├── oauth.py             # OAuth endpoints
│   │       └── health.py            # GET /health
│   │
│   ├── core/
│   │   ├── __init__.py
│   │   ├── config.py                # Settings management
│   │   ├── account.py               # Account model & manager
│   │   └── exceptions.py            # Custom exceptions
│   │
│   ├── services/
│   │   ├── __init__.py
│   │   ├── oauth.py                 # OAuth service with refresh
│   │   └── claude_api.py            # Claude API client
│   │
│   ├── models/
│   │   ├── __init__.py
│   │   ├── message.py               # Message request/response models
│   │   └── account.py               # Account data model
│   │
│   ├── middleware/
│   │   ├── __init__.py
│   │   └── auth.py                  # API key validation
│   │
│   └── utils/
│       ├── __init__.py
│       ├── logger.py                # Basic logging
│       └── retry.py                 # Retry logic
│
├── tests/
│   ├── __init__.py
│   ├── test_oauth.py
│   ├── test_messages.py
│   └── test_account.py
│
├── .env.example                     # Example environment variables
├── .gitignore
├── pyproject.toml                   # Poetry dependencies
├── README.md                        # Setup and usage guide
└── MVP_SPEC.md                      # This file
```

---

## Data Models

### Account Model

```python
from pydantic import BaseModel
from typing import Optional

class Account(BaseModel):
    organization_uuid: str
    access_token: str
    refresh_token: str
    expires_at: int  # Unix timestamp
    status: str = "valid"  # "valid" | "invalid"
    created_at: int  # Unix timestamp
    updated_at: int  # Unix timestamp
```

**Storage**: `~/.clove/data/account.json`

```json
{
  "organization_uuid": "org_abc123...",
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rt_abc123...",
  "expires_at": 1730000000,
  "status": "valid",
  "created_at": 1729900000,
  "updated_at": 1729950000
}
```

### Message Request Model

```python
from pydantic import BaseModel
from typing import List, Optional

class MessageContent(BaseModel):
    type: str = "text"
    text: str

class Message(BaseModel):
    role: str  # "user" | "assistant"
    content: str | List[MessageContent]

class MessageRequest(BaseModel):
    model: str = "claude-opus-4-20250514"
    messages: List[Message]
    max_tokens: int = 4096
    stream: bool = False
    temperature: Optional[float] = 1.0
    top_p: Optional[float] = None
    top_k: Optional[int] = None
```

### Message Response Model

```python
from pydantic import BaseModel
from typing import List

class Usage(BaseModel):
    input_tokens: int
    output_tokens: int

class ContentBlock(BaseModel):
    type: str = "text"
    text: str

class MessageResponse(BaseModel):
    id: str
    type: str = "message"
    role: str = "assistant"
    content: List[ContentBlock]
    model: str
    stop_reason: str
    usage: Usage
```

---

## Configuration

### Environment Variables

```bash
# Server Configuration
HOST=0.0.0.0
PORT=5201

# Authentication
API_KEY=your-secret-api-key

# OAuth Configuration
OAUTH_CLIENT_ID=your-claude-oauth-client-id
OAUTH_AUTHORIZE_URL=https://claude.ai/oauth/authorize
OAUTH_TOKEN_URL=https://api.claude.ai/oauth/token
OAUTH_REDIRECT_URI=http://localhost:5201/oauth/callback

# Claude API
CLAUDE_API_BASEURL=https://api.claude.ai

# Data Storage
DATA_FOLDER=~/.clove/data

# Retry Configuration
MAX_RETRIES=3
RETRY_DELAY=1

# Logging
LOG_LEVEL=INFO
```

### config.json

```json
{
  "host": "0.0.0.0",
  "port": 5201,
  "api_key": "your-secret-api-key",
  "oauth": {
    "client_id": "your-claude-oauth-client-id",
    "authorize_url": "https://claude.ai/oauth/authorize",
    "token_url": "https://api.claude.ai/oauth/token",
    "redirect_uri": "http://localhost:5201/oauth/callback"
  },
  "claude_api_baseurl": "https://api.claude.ai",
  "data_folder": "~/.clove/data",
  "max_retries": 3,
  "retry_delay": 1,
  "log_level": "INFO"
}
```

---

## Authentication Flow

### Initial OAuth Setup

1. **User requests authorization URL**

   ```bash
   curl http://localhost:5201/oauth/authorize
   ```

2. **System generates PKCE parameters**

   - code_verifier (random string)
   - code_challenge (SHA256 hash of verifier)
   - state (random string for CSRF protection)

3. **User visits authorization URL**

   ```
   https://claude.ai/oauth/authorize?
     client_id=YOUR_CLIENT_ID&
     response_type=code&
     redirect_uri=http://localhost:5201/oauth/callback&
     state=RANDOM_STATE&
     code_challenge=CHALLENGE&
     code_challenge_method=S256
   ```

4. **User authorizes and gets redirected**

   ```
   http://localhost:5201/oauth/callback?code=AUTH_CODE&state=STATE
   ```

5. **System exchanges code for tokens**

   ```http
   POST https://api.claude.ai/oauth/token
   Content-Type: application/x-www-form-urlencoded

   grant_type=authorization_code&
   code=AUTH_CODE&
   client_id=YOUR_CLIENT_ID&
   redirect_uri=http://localhost:5201/oauth/callback&
   code_verifier=VERIFIER
   ```

6. **Receive tokens**

   ```json
   {
     "access_token": "eyJhbGciOiJIUzI1NiIs...",
     "refresh_token": "rt_abc123...",
     "token_type": "Bearer",
     "expires_in": 3600
   }
   ```

7. **Fetch organization UUID**

   ```http
   GET https://api.claude.ai/v1/organizations
   Authorization: Bearer ACCESS_TOKEN
   ```

8. **Save account to storage**

### Token Refresh Flow

1. **Before each API request, check token expiry**

   ```python
   if time.time() >= account.expires_at - 60:  # 60s buffer
       await refresh_token()
   ```

2. **Refresh token request**

   ```http
   POST https://api.claude.ai/oauth/token
   Content-Type: application/x-www-form-urlencoded

   grant_type=refresh_token&
   refresh_token=REFRESH_TOKEN&
   client_id=YOUR_CLIENT_ID
   ```

3. **Receive new tokens**

   ```json
   {
     "access_token": "eyJhbGciOiJIUzI1NiIs...",
     "refresh_token": "rt_xyz789...",
     "token_type": "Bearer",
     "expires_in": 3600
   }
   ```

4. **Update account with new tokens**

   ```python
   account.access_token = new_access_token
   account.refresh_token = new_refresh_token
   account.expires_at = time.time() + expires_in
   account.updated_at = time.time()
   await account_manager.save_account(account)
   ```

5. **Proceed with original API request**

---

## Error Handling

### Error Types

```python
class CloveException(Exception):
    """Base exception for all Clove errors"""
    def __init__(self, message: str, code: int = 500):
        self.message = message
        self.code = code
        super().__init__(self.message)

class OAuthError(CloveException):
    """OAuth-related errors"""
    pass

class TokenExpiredError(OAuthError):
    """Access token has expired"""
    pass

class RefreshTokenError(OAuthError):
    """Failed to refresh token"""
    pass

class APIError(CloveException):
    """Claude API errors"""
    pass

class AuthenticationError(CloveException):
    """Authentication errors"""
    pass

class ValidationError(CloveException):
    """Request validation errors"""
    pass
```

### Error Responses

```json
{
  "error": {
    "type": "oauth_error",
    "message": "Failed to refresh access token",
    "code": 401
  }
}
```

### Retry Logic

```python
from tenacity import retry, stop_after_attempt, wait_fixed

@retry(
    stop=stop_after_attempt(3),
    wait=wait_fixed(1),
    retry=retry_if_exception_type(APIError)
)
async def send_message_with_retry(request: MessageRequest):
    return await claude_api.send_message(request)
```

---

## Testing Strategy

### Unit Tests

```python
# tests/test_oauth.py
async def test_token_exchange():
    """Test OAuth code exchange"""
    pass

async def test_token_refresh():
    """Test token refresh flow"""
    pass

async def test_token_expiry_check():
    """Test token expiry detection"""
    pass

# tests/test_account.py
async def test_save_account():
    """Test account persistence"""
    pass

async def test_load_account():
    """Test account loading"""
    pass

async def test_auto_refresh():
    """Test automatic token refresh"""
    pass

# tests/test_messages.py
async def test_send_message():
    """Test message sending"""
    pass

async def test_streaming_response():
    """Test streaming mode"""
    pass

async def test_non_streaming_response():
    """Test non-streaming mode"""
    pass
```

### Integration Tests

```python
async def test_end_to_end_flow():
    """Test complete OAuth + message flow"""
    # 1. Get authorization URL
    # 2. Exchange code for tokens
    # 3. Send message
    # 4. Verify response
    pass

async def test_token_refresh_integration():
    """Test token refresh during message sending"""
    # 1. Set token to expired
    # 2. Send message
    # 3. Verify auto-refresh occurred
    # 4. Verify message sent successfully
    pass
```

### Manual Testing Checklist

- [ ] OAuth authorization flow works
- [ ] Token exchange returns valid tokens
- [ ] Account is saved to disk
- [ ] Message sending works (non-streaming)
- [ ] Message sending works (streaming)
- [ ] Token refresh works automatically
- [ ] Expired tokens are refreshed before API calls
- [ ] API key validation works
- [ ] Health check returns correct status
- [ ] Errors are handled gracefully
- [ ] Retry logic works on failures

---

## Future Enhancements

### Phase 2 Features (Post-MVP)

1. **Multi-Account Support**

   - Load balancing
   - Account pool management
   - Automatic failover

2. **Image Support**

   - Base64 images
   - URL images
   - Image preprocessing

3. **Tool Calling**

   - Tool use support
   - Tool result handling
   - Async tool tracking

4. **Prompt Caching**

   - Cache management
   - Account affinity
   - Cache timeout

5. **Admin Interface**

   - Web dashboard
   - Account management UI
   - Statistics dashboard

6. **Advanced Features**

   - Token counting
   - Stop sequences
   - Extended thinking
   - Web search

7. **Web Proxy Mode**

   - Cookie authentication
   - Fallback to web.claude.ai
   - Dual-mode operation

8. **Production Features**
   - Rate limiting
   - Metrics/monitoring
   - Advanced logging
   - Database backend option

---

## Success Metrics

### Technical Metrics

- [ ] Response time < 2s (p95)
- [ ] Token refresh success rate > 99%
- [ ] API uptime > 99.9%
- [ ] Error rate < 1%
- [ ] Test coverage > 80%

### User Experience Metrics

- [ ] Setup time < 10 minutes
- [ ] Documentation clarity score > 4/5
- [ ] Zero-config for basic usage
- [ ] Clear error messages

---

## Dependencies

### Core Dependencies

```toml
[tool.poetry.dependencies]
python = "^3.11"
fastapi = "^0.104.0"
uvicorn = "^0.24.0"
httpx = "^0.25.0"
pydantic = "^2.4.0"
pydantic-settings = "^2.0.0"
tenacity = "^8.2.0"
python-dotenv = "^1.0.0"

[tool.poetry.group.dev.dependencies]
pytest = "^7.4.0"
pytest-asyncio = "^0.21.0"
black = "^23.10.0"
ruff = "^0.1.0"
```

---

## Getting Started (Quick Start Guide)

### 1. Installation

```bash
git clone https://github.com/yourusername/clove-mvp.git
cd clove-mvp
poetry install
```

### 2. Configuration

```bash
cp .env.example .env
# Edit .env with your settings
```

### 3. OAuth Setup

```bash
# Start the server
poetry run uvicorn app.main:app --host 0.0.0.0 --port 5201

# Get authorization URL
curl http://localhost:5201/oauth/authorize

# Visit the URL in browser, authorize, and complete callback
```

### 4. Send Your First Message

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

---

## Conclusion

This MVP specification defines a focused, achievable version of Clove that provides core OAuth authentication with refresh token support and basic message sending capabilities. The implementation can be completed in approximately 5 days and provides a solid foundation for future enhancements.

**Key Differentiators**:

- Full OAuth 2.0 with PKCE implementation
- **Automatic token refresh before expiry**
- OpenAI-compatible API
- Production-ready error handling
- Simple deployment and configuration

**Next Steps**:

1. Review and approve this specification
2. Set up development environment
3. Begin Phase 1 implementation
4. Iterate based on testing feedback

---

**Document Version**: 1.0.0
**Last Updated**: 2025-10-25
**Target Completion**: 5 days from start
