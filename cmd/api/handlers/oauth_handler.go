package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"claude-proxy/modules/proxy/application/dto"
	"claude-proxy/modules/proxy/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"
	"claude-proxy/pkg/oauth"
)

// OAuthHandler handles OAuth-related endpoints
type OAuthHandler struct {
	oauthService  *oauth.Service
	accountRepo   interfaces.AccountRepository
	accountSvc    interfaces.AccountService
	claudeBaseURL string
	challenges    map[string]*oauth.PKCEChallenge // state -> challenge
	challengesMu  sync.Mutex
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthService *oauth.Service,
	accountRepo interfaces.AccountRepository,
	accountSvc interfaces.AccountService,
	claudeBaseURL string,
) *OAuthHandler {
	return &OAuthHandler{
		oauthService:  oauthService,
		accountRepo:   accountRepo,
		accountSvc:    accountSvc,
		claudeBaseURL: claudeBaseURL,
		challenges:    make(map[string]*oauth.PKCEChallenge),
	}
}

// GetAuthorizeURL generates and returns the OAuth authorization URL with PKCE challenge
// GET /oauth/authorize?org_id=xxx (org_id is optional)
func (h *OAuthHandler) GetAuthorizeURL(c *gin.Context) {
	// Get optional organization ID from query parameter
	orgID := c.Query("org_id")

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
	authURL := h.oauthService.BuildAuthorizationURL(challenge, orgID)

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

	// Exchange code for tokens
	tokenResp, err := h.oauthService.ExchangeCodeForToken(ctx, req.Code, req.CodeVerifier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": fmt.Sprintf("Failed to exchange code for token: %v", err),
			},
		})
		return
	}

	// Get organization UUID (use provided or fetch from API)
	var orgUUID string
	if req.OrgID != "" {
		orgUUID = req.OrgID
	} else {
		orgUUID, err = h.oauthService.GetOrganizationUUID(ctx, tokenResp.AccessToken, h.claudeBaseURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"type":    "oauth_error",
					"message": fmt.Sprintf("Failed to get organization UUID: %v", err),
				},
			})
			return
		}
	}

	// Create account entity
	now := time.Now()
	acc := &entities.Account{
		ID:               uuid.New().String(),
		Name:             req.Name,
		OrganizationUUID: orgUUID,
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		ExpiresAt:        time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Status:           entities.AccountStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Save to repository
	if err := h.accountRepo.Create(ctx, acc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "oauth_error",
				"message": fmt.Sprintf("Failed to save account: %v", err),
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

// HandleCallback handles the OAuth callback (returns code and state as JSON)
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

	// Return code and state for the frontend to handle
	c.JSON(http.StatusOK, gin.H{
		"code":  code,
		"state": state,
	})
}
