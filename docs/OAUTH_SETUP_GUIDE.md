# OAuth Setup Guide - Step by Step

This guide walks you through setting up OAuth authentication for Claude Proxy to access the Claude API.

## Prerequisites

1. **Claude OAuth Client ID** - Obtain from Claude/Anthropic
2. **Running Claude Proxy Server** - Backend must be running
3. **Browser** - To authorize with Claude

---

## Configuration

### 1. Update Configuration File

Edit your `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 5201

auth:
  api_key: "your-secret-api-key-here"  # Protect your /v1/messages endpoint

oauth:
  client_id: "your-claude-oauth-client-id-here"  # Required!
  authorize_url: "https://claude.ai/oauth/authorize"
  token_url: "https://api.claude.ai/oauth/token"
  redirect_uri: "http://localhost:5201/oauth/callback"

claude:
  base_url: "https://api.claude.ai"

storage:
  data_folder: "~/.claude-proxy/data"  # Tokens stored here as account.json

retry:
  max_retries: 3
  retry_delay: 1s

logger:
  level: "info"
```

### 2. Start the Server

```bash
# Using binary
./bin/claude-proxy

# Or using Go
go run .

# Or using Make
make be
```

You should see:
```
INFO Starting Claude Proxy API Server port=5201
INFO API Endpoints:
INFO   OAuth:
INFO     GET  /oauth/authorize - Generate OAuth URL (returns state + code_verifier)
INFO     POST /oauth/exchange  - Exchange code for tokens (manual flow)
INFO     GET  /oauth/callback  - OAuth callback page (shows code to copy)
```

---

## OAuth Setup - Two Methods

You can set up OAuth using either:
1. **Frontend UI** (Recommended - easier)
2. **Command Line** (Manual - for advanced users)

---

## Method 1: Frontend UI (Recommended)

### Step 1: Access OAuth Setup Page

Open your browser and navigate to:
- Public access: `http://localhost:5201/oauth/setup`
- Admin dashboard: `http://localhost:5201/admin/oauth-setup` (requires login)

### Step 2: (Optional) Enter Organization ID

If you have a specific Claude organization you want to use:
- Enter the Organization ID (format: `org_...`)
- If left blank, it will be automatically fetched from your account

### Step 3: Generate OAuth URL

Click **"Generate OAuth URL"** button

The UI will:
- Generate a PKCE code challenge
- Create a unique state parameter
- Store the code_verifier securely
- Display the authorization URL

### Step 4: Authorize with Claude

Click the **"Connect to Claude"** button

This opens a new tab with Claude's authorization page where you:
1. Log in to your Claude account (if not already logged in)
2. Review the permissions requested
3. Click **"Authorize"** to grant access
4. You'll be redirected to: `http://localhost:5201/oauth/callback?code=AUTH_CODE&state=STATE`

### Step 5: Copy Authorization Code

From the callback URL in your browser's address bar, copy the **code** parameter

Example URL:
```
http://localhost:5201/oauth/callback?code=def502007a8c...&state=abc123...
```

Copy only the value after `code=` (everything before `&state=`)

### Step 6: Complete Setup

Back in the OAuth Setup UI:
1. Paste the authorization code into the input field
2. Click **"Complete Setup"** button

The system will:
- Validate the state matches
- Exchange the code for access and refresh tokens
- Fetch your organization UUID (if not provided)
- Save everything to `~/.claude-proxy/data/account.json`
- Display success message

### Step 7: Success!

You'll see:
- ✅ OAuth Setup Complete
- Your Organization UUID
- Token expiry time
- "Tokens will be automatically refreshed before expiry"

Click **"Go to Dashboard"** to start using Claude API!

---

## Method 2: Command Line (Manual)

### Step 1: Generate Authorization URL

```bash
curl http://localhost:5201/oauth/authorize
```

Response:
```json
{
  "authorization_url": "https://claude.ai/oauth/authorize?client_id=...&response_type=code&redirect_uri=http://localhost:5201/oauth/callback&state=abc123...&code_challenge=xyz789...&code_challenge_method=S256",
  "state": "abc123...",
  "code_verifier": "verifier456..."
}
```

**Save the `state` and `code_verifier` - you'll need them later!**

### Step 2: Visit Authorization URL

Copy the `authorization_url` and paste it into your browser.

1. Log in to Claude (if needed)
2. Review permissions
3. Click **"Authorize"**
4. You'll be redirected to: `http://localhost:5201/oauth/callback?code=AUTH_CODE&state=STATE`

### Step 3: Copy Authorization Code

From the callback URL, copy the value of the `code` parameter.

### Step 4: Exchange Code for Tokens

```bash
curl -X POST http://localhost:5201/oauth/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "code": "YOUR_AUTH_CODE_HERE",
    "state": "STATE_FROM_STEP_1",
    "code_verifier": "VERIFIER_FROM_STEP_1"
  }'
```

**With Organization ID (optional):**
```bash
curl -X POST http://localhost:5201/oauth/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "code": "YOUR_AUTH_CODE_HERE",
    "state": "STATE_FROM_STEP_1",
    "code_verifier": "VERIFIER_FROM_STEP_1",
    "org_id": "org_abc123..."
  }'
```

Success Response:
```json
{
  "success": true,
  "message": "Account configured successfully",
  "organization_uuid": "org_abc123...",
  "expires_at": 1730000000
}
```

### Step 5: Verify Setup

Check account status:
```bash
curl http://localhost:5201/health
```

Response:
```json
{
  "status": "ok",
  "account_valid": true,
  "expires_at": 1730000000,
  "organization_uuid": "org_abc123..."
}
```

---

## Test Your Setup

### Send a Test Message

```bash
curl -X POST http://localhost:5201/v1/messages \
  -H "X-API-Key: your-secret-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-20250514",
    "messages": [
      {
        "role": "user",
        "content": "Hello! Can you confirm you are working?"
      }
    ],
    "max_tokens": 1024
  }'
```

### Test Streaming

```bash
curl -X POST http://localhost:5201/v1/messages \
  -H "X-API-Key: your-secret-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-20250514",
    "messages": [
      {
        "role": "user",
        "content": "Count to 5 slowly"
      }
    ],
    "max_tokens": 1024,
    "stream": true
  }'
```

You should see Server-Sent Events (SSE) streaming back:
```
event: message_start
data: {"type":"message_start",...}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"text":"1"}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"text":"... "}}
```

---

## Token Management

### Automatic Token Refresh

Claude Proxy automatically refreshes your access token:
- Checks expiry before each API call
- Refreshes if token expires in less than 60 seconds
- Updates stored tokens in `~/.claude-proxy/data/account.json`
- No manual intervention needed!

### Account Storage Location

Your account data is stored at:
```
~/.claude-proxy/data/account.json
```

File contents:
```json
{
  "organization_uuid": "org_abc123...",
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rt_xyz789...",
  "expires_at": 1730000000,
  "status": "valid",
  "created_at": 1729900000,
  "updated_at": 1729950000
}
```

**Security Note:** This file contains sensitive tokens!
- File permissions: `0600` (read/write for owner only)
- Keep this file secure
- Never commit to git
- Back up securely if needed

### Re-authorize (Reset OAuth)

To set up a new account or re-authorize:

1. Delete the account file:
   ```bash
   rm ~/.claude-proxy/data/account.json
   ```

2. Follow the OAuth setup steps again (Method 1 or 2)

---

## Troubleshooting

### Error: "Invalid or expired state"

**Cause:** State parameter mismatch or timeout (states expire after 10 minutes)

**Solution:**
1. Generate a new authorization URL (start from Step 1)
2. Complete the flow within 10 minutes
3. Don't refresh the page during the process

### Error: "Failed to exchange code for token"

**Possible causes:**
- Authorization code already used (codes are single-use)
- Code expired (codes expire quickly, usually 5-10 minutes)
- Wrong code_verifier

**Solution:**
1. Start over from Step 1
2. Don't reuse old authorization codes
3. Complete the exchange quickly after authorization

### Error: "Failed to get valid token"

**Cause:** No account configured or refresh token invalid

**Solution:**
1. Check if `~/.claude-proxy/data/account.json` exists
2. Verify account status: `curl http://localhost:5201/health`
3. If invalid, re-run OAuth setup

### Error: "No account configured"

**Cause:** OAuth setup not completed

**Solution:** Complete OAuth setup using Method 1 or 2

### Account Status Shows "invalid"

**Cause:** Refresh token expired or revoked

**Solution:**
1. Delete account: `rm ~/.claude-proxy/data/account.json`
2. Re-run OAuth setup
3. Authorize with Claude again

### Error: "Failed to refresh token"

**Cause:** Refresh token expired or account access revoked

**Solution:**
1. Check Claude account status
2. Re-authorize through OAuth setup
3. Ensure API credentials are still valid

---

## Security Best Practices

1. **Protect your API Key** (`auth.api_key` in config)
   - Use strong, random keys
   - Don't share in public repos
   - Rotate regularly

2. **Secure Token Storage**
   - Keep `~/.claude-proxy/data/account.json` secure
   - File has `0600` permissions (owner-only access)
   - Back up securely if needed

3. **HTTPS in Production**
   - Use HTTPS for production deployments
   - Update `redirect_uri` to use `https://`
   - Configure SSL/TLS certificates

4. **Network Security**
   - Run behind a reverse proxy (nginx, Caddy)
   - Use firewall rules to restrict access
   - Consider VPN for sensitive environments

---

## API Endpoints Reference

### OAuth Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/oauth/authorize` | Generate OAuth URL + PKCE challenge | No |
| POST | `/oauth/exchange` | Exchange code for tokens | No |
| GET | `/oauth/callback` | OAuth callback (returns code/state) | No |

### Claude API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/v1/messages` | Send message to Claude | Yes (X-API-Key) |

### Health Check

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Check server and account status | No |

---

## What's Next?

After successful OAuth setup:

1. **Integrate with your application** - Use the `/v1/messages` endpoint
2. **Monitor token refresh** - Check logs for automatic refresh operations
3. **Set up monitoring** - Monitor the `/health` endpoint
4. **Configure production** - Set up HTTPS, reverse proxy, etc.

---

## Need Help?

- Check server logs for detailed error messages
- Verify configuration in `config.yaml`
- Ensure Claude OAuth client ID is valid
- Test with `curl` before integrating with your app

**Pro Tip:** The frontend UI (Method 1) is the easiest way to set up OAuth. It handles all the complexity for you with a clean step-by-step wizard! 🚀
