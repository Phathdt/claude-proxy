# Claude Proxy - Feature Documentation

> Comprehensive feature list for Claude Proxy reverse proxy v0.1.0

## Table of Contents

- [Project Overview](#project-overview)
- [Core Proxy Functionality](#core-proxy-functionality)
- [Account Management](#account-management)
- [API Endpoints & Messaging](#api-endpoints--messaging)
- [Advanced API Features](#advanced-api-features)
- [Response Processing Pipeline](#response-processing-pipeline)
- [Session Management](#session-management)
- [Configuration & Settings](#configuration--settings)
- [Authentication & Security](#authentication--security)
- [Admin Management Interface](#admin-management-interface)
- [Internationalization](#internationalization)
- [Error Handling & Logging](#error-handling--logging)
- [Utility Features](#utility-features)
- [Static Files & Frontend](#static-files--frontend)
- [Deployment & Startup](#deployment--startup)
- [Content Handling](#content-handling)
- [Edge Cases & Special Features](#edge-cases--special-features)
- [Database Models & Data Structures](#database-models--data-structures)
- [Configuration Examples](#configuration-examples)
- [Unique/Standout Capabilities](#uniquestandout-capabilities)

---

## Project Overview

**Type**: Go Gin-based Claude reverse proxy with React admin dashboard
**Version**: 0.1.0
**Go Support**: 1.21+
**Primary Purpose**: OAuth2-based proxy for multi-account Claude API access with automatic token refresh

---

## Core Proxy Functionality

### OAuth Authentication

**Files**: `/app/services/oauth.py`, `/app/core/account.py`

- OAuth 2.0 authentication with PKCE support for Claude API access
- Automatic OAuth token exchange from authorization codes
- OAuth token refresh capability
- Support for multiple authentication types (OAuth-only, Cookie-only, Both)
- Organization UUID discovery from Claude.ai

### Dual Mode Operation

**Files**: `/app/processors/claude_ai/claude_api_processor.py`, `/app/processors/claude_ai/claude_web_processor.py`

- **OAuth Mode**: Direct Claude API access using OAuth tokens
- **Web Proxy Mode**: Claude.ai web interface emulation for fallback
- Intelligent mode switching based on account capabilities

### Cookie-Based Authentication

- Session-based cookie management for Claude.ai
- Masked cookie storage and display (only first 20 characters shown)
- Cookie validation and refresh mechanisms

---

## Account Management

### Account Management System

**Files**: `/app/services/account.py`, `/app/core/account.py`

**Features**:

- Multi-account support with load balancing
- Organization UUID tracking
- Account status tracking (VALID, INVALID, RATE_LIMITED)
- Account capability detection (Pro, Max, Free tier identification)
- Last used timestamp tracking
- Rate limit reset time tracking
- Session per account limiting (max 3 concurrent sessions configurable)
- Singleton pattern implementation for thread-safe access
- Persistent account storage (JSON-based)
- Automatic account status recovery

### Account API Endpoints

**File**: `/app/api/routes/accounts.py`

- `GET /api/admin/accounts` - List all accounts
- `GET /api/admin/accounts/{organization_uuid}` - Get specific account
- `POST /api/admin/accounts` - Create new account
- `PUT /api/admin/accounts/{organization_uuid}` - Update account
- `DELETE /api/admin/accounts/{organization_uuid}` - Delete account
- `POST /api/admin/accounts/oauth/exchange` - OAuth code exchange

---

## API Endpoints & Messaging

### Claude Messages API Compatibility

**File**: `/app/api/routes/claude.py`

- `POST /v1/messages` - Create messages (streaming & non-streaming)
- Full OpenAI-compatible API interface
- Retry mechanism with exponential backoff
- Request validation and error handling

### Message Processing Features

**Files**: `/app/utils/messages.py`, `/app/models/claude.py`

Support for complex message structures with multiple content types:

- Text content
- Image content (base64, URL, file-based)
- Thinking content (extended thinking/chain of thought)
- Tool use content
- Tool result content
- Server tool use (web search)
- Web search results

### Supported Models

- claude-opus-4-20250514 (default)
- All Claude variants (Opus, Sonnet, Haiku)
- Dynamic model injection based on request

---

## Advanced API Features

### Tool/Function Calling

**Files**: `/app/processors/claude_ai/tool_call_event_processor.py`, `/app/services/tool_call.py`

- Full tool use support with async tool call tracking
- Tool call state management with timeout (configurable: 300s)
- Tool result handling and session resumption
- Parallel tool use support
- Tool call cleanup with automatic expiration

### Stop Sequences

**File**: `/app/processors/claude_ai/stop_sequences_processor.py`

- Custom stop sequence detection and enforcement
- Stop sequence interruption of streaming responses
- MessageDelta event injection on stop sequence detection

### Token Counting

**File**: `/app/processors/claude_ai/token_counter_processor.py`

- Tiktoken-based token estimation (cl100k_base encoding)
- Input and output token counting
- Cache creation and read token tracking
- Automatic token usage injection into response events

### Prompt Caching

**File**: `/app/services/cache.py`

- Cache checkpoint management
- Account affinity based on cached prompts
- Cache timeout management (configurable: 300s)
- Automatic cache cleanup with configurable intervals

### Extended Thinking Support

- Native thinking content type support
- Budget token configuration for thinking blocks
- Thinking delta event streaming

### Web Search Integration

- Server tool use content support
- Web search result handling
- Web search request tracking and usage statistics

---

## Response Processing Pipeline

### Processing Pipeline Architecture

**Files**: `/app/processors/claude_ai/pipeline.py`, `/app/processors/base.py`

**12-stage request processing pipeline**:

1. **TestMessageProcessor** - Handles test message responses
2. **ToolResultProcessor** - Resumes paused sessions from tool results
3. **ClaudeAPIProcessor** - OAuth-based API requests
4. **ClaudeWebProcessor** - Web proxy request building
5. **EventParsingProcessor** - SSE stream parsing
6. **ModelInjectorProcessor** - Model information injection
7. **StopSequencesProcessor** - Stop sequence handling
8. **ToolCallEventProcessor** - Tool call event processing
9. **MessageCollectorProcessor** - Real-time message collection
10. **TokenCounterProcessor** - Token usage estimation
11. **StreamingResponseProcessor** - SSE response streaming
12. **NonStreamingResponseProcessor** - JSON response building

### Streaming & Response Formats

**Files**: `/app/processors/claude_ai/streaming_response_processor.py`, `/app/processors/claude_ai/non_streaming_response_processor.py`

- Server-Sent Events (SSE) streaming support
- JSON non-streaming response format
- Event serialization with proper formatting
- Error event handling and propagation

### Event Processing

**Files**: `/app/services/event_processing/event_parser.py`, `/app/services/event_processing/event_serializer.py`

- Complete streaming event model coverage
- Event type validation and routing
- Delta-based streaming for efficient data transfer
- Event serialization to standard formats

---

## Session Management

### Session Management System

**File**: `/app/services/session.py`

- Session creation and retrieval
- Claude web session management with account binding
- Session timeout (configurable: 300s)
- Automatic session cleanup
- Session expiration detection
- Concurrent session limiting per cookie

### Session Features

- Per-session request/response tracking
- Conversation state preservation
- Async session lifecycle management

---

## Configuration & Settings

### Settings Management

**File**: `/app/api/routes/settings.py`

- `GET /api/admin/settings` - Get all settings
- `PUT /api/admin/settings` - Update settings
- JSON config file support (`~/.claude-proxy/data/config.json`)
- Environment variable override support
- Dynamic settings updates without restart

### Configuration Options

**File**: `/app/core/config.py`

- Host and port settings (default: 0.0.0.0:5201)
- API key management (standard + admin keys)
- Proxy URL configuration
- Claude URLs (API and web)
- Session timeouts and cleanup intervals
- Tool call timeout management
- Cache timeout and cleanup settings
- OAuth configuration
- Request timeout and retry settings
- Logging configuration with file rotation
- Content processing options
- Custom system prompt
- Padding token configuration
- External image download permissions
- Conversation preservation mode
- Multiple language support

### Data Storage

- Persistent data folder (`~/.claude-proxy/data/`)
- JSON-based account storage
- JSON-based configuration storage
- No-filesystem mode support for serverless environments
- Automatic data folder creation

---

## Authentication & Security

### API Key Management

**File**: `/app/dependencies/auth.py`

- Support for `X-API-Key` header
- Support for `Authorization: Bearer` header
- Separate user API keys and admin API keys
- Temporary admin key generation if none configured
- Key validation on all admin endpoints
- Key validation on all standard endpoints

### Admin Interface Protection

- Admin-only endpoints require admin API key
- User endpoints require standard API key
- Automatic temporary key generation with warning
- Multi-level API key validation

---

## Admin Management Interface

### Statistics Endpoint

**File**: `/app/api/routes/statistics.py`

- `GET /api/admin/statistics` - System statistics
- Account statistics (total, valid, rate-limited, invalid)
- Active session count
- System health status

### Health Check

- `GET /health` - Health status endpoint
- Status based on available valid accounts

---

## Internationalization

### Multi-Language Support

**File**: `/app/services/i18n.py`

- Translation file loading from locales directory
- Support for nested translation keys (dot notation)
- Message interpolation with context variables
- Accept-Language header parsing
- Default language fallback
- JSON-based translation files

### Locales Directory

**Location**: `/app/locales/`

- Support for multiple language packs
- Dynamic translation loading

---

## Error Handling & Logging

### Comprehensive Exception System

**File**: `/app/core/exceptions.py`

- 25+ custom exception types
- Structured error codes (6-digit format)
- Error message localization
- Retryable error detection
- Context-aware error information

**Exception types**:

- Authentication errors (Claude, OAuth, Cookie)
- Authorization errors
- Rate limiting errors
- Account management errors
- API errors (HTTP errors, stream errors)
- Configuration errors
- Validation errors
- Tool calling errors

### Error Handler

**File**: `/app/core/error_handler.py`

- Centralized error response formatting
- Language-aware error messages
- Standard error response format
- Error logging with context

### Logging System

**File**: `/app/utils/logger.py`

- Loguru-based structured logging
- Log level configuration
- File rotation support (configurable)
- Log retention policies
- Log compression (ZIP format)
- Async logging support

---

## Utility Features

### Retry Mechanism

**File**: `/app/utils/retry.py`

- Tenacity-based retry logic
- Configurable retry attempts (default: 3)
- Configurable retry interval (default: 1 second)
- Retry-eligible error detection
- Before-sleep logging

### Message Processing Utilities

**File**: `/app/utils/messages.py`

- Content block type handling
- Message format conversion
- Attachment processing
- Image content handling

### HTTP Client Abstraction

**File**: `/app/core/http_client.py`

**Multiple backend support**:

- rnet (preferred for web proxy)
- curl_cffi (Cloudflare bypass)
- httpx (standard HTTP)

**Features**:

- Automatic fallback between backends
- Impersonation support (Chrome user agent)
- Proxy support
- Timeout configuration
- Response streaming

---

## Static Files & Frontend

### Static File Serving

**File**: `/app/core/static.py`

- Asset serving (CSS, JS, images)
- Single Page Application (SPA) support
- Automatic index.html routing
- Frontend build output serving

---

## Deployment & Startup

### Application Lifecycle

**File**: `/app/main.py`

- FastAPI application setup
- CORS middleware configuration (allow all origins)
- Lifespan context manager for startup/shutdown
- Automatic account loading on startup
- Background task management
- Graceful shutdown handling

### Background Tasks

- Account status check task
- Session cleanup task
- Tool call cleanup task
- Cache cleanup task

### CORS Configuration

- Allow all origins
- Allow all methods and headers
- Credentials support

---

## Content Handling

### Image Support

**Types**: JPEG, PNG, GIF, WebP

**Multiple source types**:

- Base64-encoded images
- URL-based images
- File-based uploads

**Features**:

- External image download support (configurable)
- Image content processing and transformation

### File Handling

- File upload support
- File UUID tracking
- Attachment extraction and processing
- File size tracking

### Text Processing

- Custom system prompts
- Role customization (Human/Assistant names)
- Text padding with tokens
- Real role preservation option

---

## Edge Cases & Special Features

### SillyTavern Compatibility

- Special test message handling for SillyTavern integration
- Tavern-specific message processor
- Compatibility layer for non-standard requests

### Rate Limiting Handling

- Automatic rate limit detection
- Rate limit reset time tracking
- Account status marking on rate limit
- Automatic recovery on reset
- Rate limit error propagation

### Organization Disabled Detection

- Organization status tracking
- Disabled organization handling
- Account invalidation on organization disable

---

## Database Models & Data Structures

### Account Model

- Organization UUID
- Cookie value
- OAuth token (access, refresh, expiry)
- Account status and auth type
- Capabilities list
- Last used timestamp
- Rate limit reset time

### Message Models

- Input and output messages
- Role-based message structure
- Complex content blocks
- Usage tracking (input/output/cache tokens)

### Streaming Event Models

- Message start/stop events
- Content block start/delta/stop events
- Message delta with usage
- Error events
- Ping events
- Tool-specific events

---

## Configuration Examples

### Environment Variables Supported

```bash
# Server Configuration
PORT=5201
HOST=0.0.0.0
DATA_FOLDER=~/.claude-proxy/data

# Authentication
ADMIN_API_KEYS=admin-key-1,admin-key-2
API_KEYS=user-key-1,user-key-2

# Claude Configuration
COOKIES=cookie1,cookie2
CLAUDE_AI_URL=https://claude.ai
CLAUDE_API_BASEURL=https://api.claude.ai

# Proxy
PROXY_URL=http://proxy:8080

# Content Customization
CUSTOM_PROMPT=Custom system prompt
CUSTOM_HUMAN_NAME=User
CUSTOM_ASSISTANT_NAME=Assistant
PADTXT_LENGTH=100
ALLOW_EXTERNAL_IMAGES=true

# Session Management
SESSION_TIMEOUT=300
SESSION_CLEANUP_INTERVAL=60
MAX_SESSIONS_PER_COOKIE=3

# Tool Call Management
TOOL_CALL_TIMEOUT=300
TOOL_CALL_CLEANUP_INTERVAL=60

# Cache Management
CACHE_TIMEOUT=300
CACHE_CLEANUP_INTERVAL=60

# OAuth Configuration
OAUTH_CLIENT_ID=your-client-id
OAUTH_AUTHORIZE_URL=https://claude.ai/oauth/authorize
OAUTH_TOKEN_URL=https://api.claude.ai/oauth/token
OAUTH_REDIRECT_URI=http://localhost:5201/oauth/callback

# Logging
LOG_LEVEL=INFO
LOG_TO_FILE=true
LOG_FILE_PATH=~/.claude-proxy/logs/claude-proxy.log
LOG_FILE_ROTATION=10 MB
```

---

## Unique/Standout Capabilities

### 1. First OAuth-supporting Claude Reverse Proxy

Native access to Claude API features including system messages and prefilling, providing superior functionality compared to web-only proxies.

### 2. Dual-Mode Operation

Seamlessly switches between OAuth (full-featured) and web proxy (fallback) based on account capabilities and availability.

### 3. Comprehensive Compatibility

Works with SillyTavern, ChatGPT clients, and any standard OpenAI-compatible API consumer.

### 4. Advanced Streaming

Server-Sent Events (SSE) based streaming with real-time token counting and stop sequence detection.

### 5. Tool Call Support

Complex tool calling with async handling, session resumption, and parallel tool use support.

### 6. Prompt Caching

Account-aware cache management for optimized requests, reducing latency and token usage.

### 7. Extended Thinking

Native support for Claude's extended thinking/chain-of-thought with configurable budget tokens.

### 8. Multi-Language Support

Full i18n implementation with localized error messages and dynamic language selection.

### 9. Admin Web Interface

Complete settings and account management without manual config file editing.

### 10. No-Filesystem Mode

Serverless/containerized deployment support with optional persistent storage.

### 11. Intelligent Load Balancing

Smart account selection based on status, capabilities, rate limits, and cache affinity.

### 12. Automatic Recovery

Self-healing account management with automatic status recovery and rate limit reset detection.

---

## Architecture Highlights

### Design Patterns

- **Singleton Pattern**: Account manager for thread-safe global access
- **Pipeline Pattern**: 12-stage request processing pipeline
- **Factory Pattern**: HTTP client backend selection
- **Strategy Pattern**: Different processors for OAuth vs Web modes
- **Observer Pattern**: Event-based streaming response handling

### Performance Optimizations

- Async/await throughout for non-blocking I/O
- Background task management for cleanup operations
- Efficient token counting with caching
- Connection pooling via HTTP client backends
- Lazy loading of translation files

### Scalability Features

- Multi-account support with load balancing
- Configurable session and timeout limits
- Background cleanup tasks prevent memory leaks
- Stateless design (except for necessary session state)
- Horizontal scaling ready (with external session store)

---

## Go Implementation Roadmap

### Features to Port from Python Version

The following features exist in the Python version but are **missing** in the current Go implementation:

#### ✅ Recently Implemented

**0. Streaming Support (SSE/Server-Sent Events)**
- **Status**: ✅ **Implemented** - Real-time SSE streaming fully supported
- **Implementation Date**: 2025-10-27
- **Features**:
  - Automatic detection of `text/event-stream` content type
  - Real-time streaming using Gin's `c.Stream()` method
  - 4KB buffer chunks for efficient streaming
  - Graceful handling of client disconnections
  - Context-aware streaming (stops on timeout/cancellation)
  - No buffering - streams directly from Claude API to client
- **Benefits Delivered**:
  - ✅ Real-time feedback for extended thinking (1-3+ minutes)
  - ✅ Lower memory usage - no full response buffering
  - ✅ Better user experience with immediate response chunks
  - ✅ Handles long generations efficiently
- **Files**: `cmd/api/handlers/proxy_handler.go:60-127`
- **Usage**: Works automatically when client sends `"stream": true` in request

**📝 Timeout Configuration (Added in v0.1.1)**
- **Config**: `server.request_timeout` (default: 5 minutes)
- **Reason**: LLM APIs need longer timeouts for:
  - Extended thinking mode (1-3+ minutes)
  - Long response generation
  - Image analysis
  - Large context processing
- **Environment Variable**: `SERVER__REQUEST_TIMEOUT=10m`
- **Files**: `config/config.go:38`, `cmd/api/providers.go:175`

#### 🔥 High Priority (Week 1)

**1. Rate Limit Detection & Recovery**
- **Status**: Not implemented
- **Description**: Track when accounts hit Claude API rate limits (429 errors)
- **Implementation**:
  - Add `RateLimitedUntil time.Time` field to Account entity
  - Add `AccountStatusRateLimited` status type
  - Auto-recover accounts when rate limit expires
  - Skip rate-limited accounts in load balancing
- **Benefit**: Prevents wasted API calls to rate-limited accounts
- **Files**: `modules/proxy/domain/entities/account.go`, `modules/proxy/application/services/proxy_service.go`

**2. Enhanced Account Status System**
- **Current**: Only `active/inactive` statuses
- **Missing**: `rate_limited`, `invalid` (OAuth tokens revoked)
- **Implementation**:
  ```go
  const (
      AccountStatusActive      AccountStatus = "active"
      AccountStatusInactive    AccountStatus = "inactive"
      AccountStatusRateLimited AccountStatus = "rate_limited"
      AccountStatusInvalid     AccountStatus = "invalid"
  )
  ```
- **Files**: `modules/proxy/domain/entities/account.go`

**3. Statistics & Health Monitoring Endpoint**
- **Status**: Not implemented
- **Endpoint**: `GET /api/admin/statistics`
- **Response**:
  ```json
  {
    "total_accounts": 5,
    "active_accounts": 3,
    "rate_limited_accounts": 1,
    "invalid_accounts": 1,
    "oldest_token_age_hours": 2.5,
    "accounts_needing_refresh": 2
  }
  ```
- **Benefit**: Real-time monitoring of account health
- **Files**: `cmd/api/handlers/statistics_handler.go`, `cmd/api/server.go`

#### ⚡ Important (Week 2)

**4. Idle Account Detection**
- **Status**: Not implemented
- **Description**: Track and alert on accounts idle for > 5 hours
- **Implementation**:
  - Add `LastUsedAt time.Time` field to Account entity
  - Add background task to check idle accounts
  - Add `IsIdleTooLong(threshold time.Duration) bool` method
- **Benefit**: Identify unused accounts and optimize token refresh
- **Files**: `modules/proxy/domain/entities/account.go`, `modules/proxy/infrastructure/jobs/idle_checker.go`

**5. Session/Request Limiting Per Account**
- **Status**: Not implemented
- **Description**: Limit concurrent requests per account (default: 3)
- **Implementation**:
  - Add `ActiveRequestCount int` field
  - Add `MaxConcurrentRequests int` field
  - Skip accounts at max capacity in load balancer
- **Benefit**: Prevent account overload and improve request distribution
- **Files**: `modules/proxy/domain/entities/account.go`, `modules/proxy/application/services/proxy_service.go`

**6. Retry Logic with Exponential Backoff**
- **Status**: Basic retry exists, needs enhancement
- **Description**: Sophisticated retry for transient API errors
- **Implementation**:
  - Exponential backoff (1s, 2s, 4s, 8s...)
  - Retry only on specific errors (429, 502, 503, 504)
  - Max 3 retries configurable
- **Benefit**: Improve reliability for transient failures
- **Files**: `pkg/retry/retry.go`

#### 📦 Nice to Have (Week 3)

**7. Account Capability Detection**
- **Status**: Not implemented
- **Description**: Detect Free/Pro/Max tier from Claude API
- **Implementation**:
  ```go
  type AccountCapability string
  const (
      CapabilityFree AccountCapability = "free"
      CapabilityPro  AccountCapability = "pro"
      CapabilityMax  AccountCapability = "max"
  )
  // Add to Account struct
  Capabilities []AccountCapability
  ```
- **Benefit**: Tier-aware load balancing
- **Files**: `modules/proxy/domain/entities/account.go`

**8. Organization UUID Validation**
- **Status**: Not implemented
- **Description**: Periodically validate organization is still active
- **Implementation**:
  - Background task to check org validity
  - Auto-deactivate accounts with disabled/deleted orgs
- **Benefit**: Detect organization changes without manual intervention
- **Files**: `modules/proxy/infrastructure/jobs/org_validator.go`

**9. Enhanced Health Check**
- **Status**: Basic `/health` exists, needs enhancement
- **Current**: Returns simple status
- **Enhancement**:
  ```json
  {
    "status": "healthy|degraded|unhealthy",
    "available_accounts": 3,
    "rate_limited_accounts": 1,
    "uptime_seconds": 3600
  }
  ```
- **Status Logic**:
  - `healthy`: ≥2 active accounts
  - `degraded`: 1 active account
  - `unhealthy`: 0 active accounts
- **Files**: `cmd/api/handlers/health_handler.go`

**10. Account Usage Metrics**
- **Status**: Not implemented
- **Description**: Track request counts and token usage
- **Implementation**:
  ```go
  // Add to Account struct
  TotalRequests int64
  TotalTokensUsed int64
  LastRequestAt time.Time
  RequestsToday int  // Reset daily
  ```
- **Benefit**: Usage analytics and billing insights
- **Files**: `modules/proxy/domain/entities/account.go`

### Quick Wins (Can Implement Today)

**1. Add `LastUsedAt` Tracking** (10 minutes)
```go
// In proxy_service.go:ProxyRequest()
account.LastUsedAt = time.Now()
s.accountRepo.Update(ctx, account)
```

**2. Basic Statistics Endpoint** (30 minutes)
```go
// cmd/api/handlers/statistics_handler.go
func (h *StatisticsHandler) GetStatistics(c *gin.Context) {
    accounts, _ := h.accountService.ListAccounts(c.Request.Context())

    stats := map[string]interface{}{
        "total": len(accounts),
        "active": countActive(accounts),
        "needing_refresh": countNeedingRefresh(accounts),
    }

    c.JSON(200, stats)
}
```

### Implementation Status Comparison

| Feature | Python Version | Go Version | Priority |
|---------|---------------|------------|----------|
| OAuth Authentication | ✅ | ✅ | - |
| Token Auto-Refresh | ✅ | ✅ | - |
| Multi-Account Load Balancing | ✅ | ✅ | - |
| **SSE Streaming** | ✅ | ✅ | **✅ Done** |
| **Configurable Timeout** | ✅ | ✅ | **✅ Done** |
| Rate Limit Detection | ✅ | ❌ | 🔥 High |
| Account Status System | ✅ (4 states) | ⚠️ (2 states) | 🔥 High |
| Statistics Endpoint | ✅ | ❌ | 🔥 High |
| Idle Account Detection | ✅ | ❌ | ⚡ Important |
| Session Limiting | ✅ | ❌ | ⚡ Important |
| Exponential Backoff | ✅ | ⚠️ Basic | ⚡ Important |
| Capability Detection | ✅ | ❌ | 📦 Nice to Have |
| Org Validation | ✅ | ❌ | 📦 Nice to Have |
| Usage Metrics | ✅ | ❌ | 📦 Nice to Have |

---

## Future Roadmap Considerations

Based on the current feature set, potential enhancements could include:

- Redis-based session storage for horizontal scaling
- Metrics and monitoring integration (Prometheus/Grafana)
- Rate limiting middleware for API consumers
- WebSocket support for real-time bidirectional communication
- Plugin system for custom processors
- Database backend option (PostgreSQL/MongoDB)
- Advanced caching strategies (Redis, Memcached)
- API versioning support
- GraphQL endpoint
- Comprehensive API documentation (OpenAPI/Swagger)

---

## License & Attribution

Claude Proxy is an open-source project. Please refer to the LICENSE file in the repository for terms and conditions.

---

**Documentation Version**: 1.0.0
**Last Updated**: 2025-10-26
**Claude Proxy Version**: 0.1.0
