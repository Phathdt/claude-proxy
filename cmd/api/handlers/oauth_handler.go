package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/interfaces"
	"claude-proxy/modules/auth/infrastructure/clients"
)

// OAuthHandler handles OAuth-related endpoints
type OAuthHandler struct {
	oauthClient   interfaces.OAuthClient
	accountSvc    interfaces.AccountService
	claudeBaseURL string
	challenges    map[string]*clients.PKCEChallenge // state -> challenge
	challengesMu  sync.Mutex
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthClient interfaces.OAuthClient,
	accountSvc interfaces.AccountService,
	claudeBaseURL string,
) *OAuthHandler {
	return &OAuthHandler{
		oauthClient:   oauthClient,
		accountSvc:    accountSvc,
		claudeBaseURL: claudeBaseURL,
		challenges:    make(map[string]*clients.PKCEChallenge),
	}
}

// GetAuthorizeURL generates and returns the OAuth authorization URL with PKCE challenge
// GET /oauth/authorize?org_id=xxx (org_id is optional)
func (h *OAuthHandler) GetAuthorizeURL(c *gin.Context) {
	// Get optional organization ID from query parameter
	orgID := c.Query("org_id")

	// Generate PKCE challenge
	challenge, err := h.oauthClient.GeneratePKCEChallenge()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": "Failed to generate OAuth challenge",
			},
		})
		return
	}

	// Store challenge for later use (when user submits the code)
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

	// Build authorization URL with organization ID if provided
	authURL := h.oauthClient.BuildAuthorizationURL(challenge, orgID)

	c.JSON(http.StatusOK, gin.H{
		"authorization_url": authURL,
		"state":             challenge.State,
		"code_verifier":     challenge.CodeVerifier,
	})
}

// ExchangeCodeRequest represents the request body for code exchange
type ExchangeCodeRequest struct {
	Name         string `json:"name"             binding:"required"` // Account name
	Code         string `json:"code"             binding:"required"`
	State        string `json:"state"            binding:"required"`
	CodeVerifier string `json:"code_verifier"    binding:"required"`
	OrgID        string `json:"org_id,omitempty"` // Optional organization ID
}

// ExchangeCode exchanges authorization code for tokens (manual flow)
// POST /oauth/exchange
func (h *OAuthHandler) ExchangeCode(c *gin.Context) {
	var req ExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	// Verify state matches a stored challenge
	h.challengesMu.Lock()
	challenge, exists := h.challenges[req.State]
	if exists {
		delete(h.challenges, req.State)
	}
	h.challengesMu.Unlock()

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": "Invalid or expired state. Please generate a new authorization URL.",
			},
		})
		return
	}

	// Verify code verifier matches
	if challenge.CodeVerifier != req.CodeVerifier {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": "Code verifier mismatch",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use AccountService to create account (handles OAuth exchange)
	acc, err := h.accountSvc.CreateAccount(ctx, req.Name, req.Code, req.CodeVerifier, req.OrgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": fmt.Sprintf("Failed to create account: %v", err),
			},
		})
		return
	}

	// Convert to response DTO
	accountResponse := dto.ToAccountResponse(acc)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account configured successfully",
		"account": accountResponse,
	})
}
