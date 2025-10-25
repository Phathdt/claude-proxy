package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"claude-proxy/pkg/account"
	"claude-proxy/pkg/oauth"
)

// OAuthHandler handles OAuth-related endpoints
type OAuthHandler struct {
	oauthService   *oauth.Service
	accountManager *account.Manager
	claudeBaseURL  string
	challenges     map[string]*oauth.PKCEChallenge // state -> challenge
	challengesMu   sync.Mutex
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *oauth.Service, accountManager *account.Manager, claudeBaseURL string) *OAuthHandler {
	return &OAuthHandler{
		oauthService:   oauthService,
		accountManager: accountManager,
		claudeBaseURL:  claudeBaseURL,
		challenges:     make(map[string]*oauth.PKCEChallenge),
	}
}

// GetAuthorizeURL generates and returns the OAuth authorization URL
// GET /oauth/authorize
func (h *OAuthHandler) GetAuthorizeURL(c *gin.Context) {
	// Generate PKCE challenge
	challenge, err := h.oauthService.GeneratePKCEChallenge()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": "Failed to generate OAuth challenge",
			},
		})
		return
	}

	// Store challenge (in production, use Redis or similar)
	h.challengesMu.Lock()
	h.challenges[challenge.State] = challenge
	h.challengesMu.Unlock()

	// Clean up old challenges after 10 minutes
	go func() {
		time.Sleep(10 * time.Minute)
		h.challengesMu.Lock()
		delete(h.challenges, challenge.State)
		h.challengesMu.Unlock()
	}()

	// Build authorization URL
	authURL := h.oauthService.BuildAuthorizationURL(challenge)

	c.JSON(http.StatusOK, gin.H{
		"authorization_url": authURL,
		"state":             challenge.State,
	})
}

// HandleCallback handles the OAuth callback
// GET /oauth/callback?code=...&state=...
func (h *OAuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": "Missing code or state parameter",
			},
		})
		return
	}

	// Retrieve challenge
	h.challengesMu.Lock()
	challenge, exists := h.challenges[state]
	if exists {
		delete(h.challenges, state)
	}
	h.challengesMu.Unlock()

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": "Invalid or expired state parameter",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Exchange code for tokens
	tokenResp, err := h.oauthService.ExchangeCodeForToken(ctx, code, challenge.CodeVerifier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": fmt.Sprintf("Failed to exchange code for token: %v", err),
			},
		})
		return
	}

	// Get organization UUID
	orgUUID, err := h.oauthService.GetOrganizationUUID(ctx, tokenResp.AccessToken, h.claudeBaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": fmt.Sprintf("Failed to get organization UUID: %v", err),
			},
		})
		return
	}

	// Save account
	acc := &account.Account{
		OrganizationUUID: orgUUID,
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		ExpiresAt:        time.Now().Unix() + int64(tokenResp.ExpiresIn),
	}

	if err := h.accountManager.SaveAccount(acc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": fmt.Sprintf("Failed to save account: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"message":           "Account configured successfully",
		"organization_uuid": orgUUID,
	})
}
