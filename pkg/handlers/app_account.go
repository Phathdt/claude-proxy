package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"claude-proxy/pkg/account"
	"claude-proxy/pkg/oauth"
)

// AppAccountHandler handles app account management endpoints
type AppAccountHandler struct {
	oauthService *oauth.Service
	multiAcctMgr *account.MultiAccountManager
	claudeBaseURL string
}

// NewAppAccountHandler creates a new app account handler
func NewAppAccountHandler(oauthService *oauth.Service, multiAcctMgr *account.MultiAccountManager, claudeBaseURL string) *AppAccountHandler {
	return &AppAccountHandler{
		oauthService:  oauthService,
		multiAcctMgr:  multiAcctMgr,
		claudeBaseURL: claudeBaseURL,
	}
}

// CreateAppAccountRequest represents the request to create an app account
type CreateAppAccountRequest struct {
	Name  string `json:"name" binding:"required"`
	OrgID string `json:"org_id"` // Optional, will be fetched if not provided
}

// CreateAppAccountResponse represents the create response
type CreateAppAccountResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
	CodeVerifier     string `json:"code_verifier"`
}

// CreateAppAccount initiates OAuth flow for a new app account
// POST /api/app-accounts
func (h *AppAccountHandler) CreateAppAccount(c *gin.Context) {
	var req CreateAppAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

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

	// Store in context for callback (in production, use Redis or similar)
	// For now, we'll rely on the frontend to store and pass back

	// Build authorization URL with organization ID if provided
	authURL := h.oauthService.BuildAuthorizationURL(challenge, req.OrgID)

	c.JSON(http.StatusOK, CreateAppAccountResponse{
		AuthorizationURL: authURL,
		State:            challenge.State,
		CodeVerifier:     challenge.CodeVerifier,
	})
}

// CompleteAppAccountRequest represents the OAuth completion request
type CompleteAppAccountRequest struct {
	Name         string `json:"name" binding:"required"`
	Code         string `json:"code" binding:"required"`
	State        string `json:"state" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"`
	OrgID        string `json:"org_id,omitempty"`
}

// CompleteAppAccount completes OAuth flow and creates the app account
// POST /api/app-accounts/complete
func (h *AppAccountHandler) CompleteAppAccount(c *gin.Context) {
	var req CompleteAppAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": fmt.Sprintf("Invalid request: %v", err),
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

	// Get organization UUID
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

	// Create app account
	appAccount, err := h.multiAcctMgr.CreateAccount(
		req.Name,
		orgUUID,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		tokenResp.ExpiresIn,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "internal_error",
				"message": fmt.Sprintf("Failed to create account: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "App account created successfully",
		"account": appAccount,
	})
}

// ListAppAccounts lists all app accounts
// GET /api/app-accounts
func (h *AppAccountHandler) ListAppAccounts(c *gin.Context) {
	accounts := h.multiAcctMgr.ListAccounts()
	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// GetAppAccount gets a single app account
// GET /api/app-accounts/:id
func (h *AppAccountHandler) GetAppAccount(c *gin.Context) {
	id := c.Param("id")

	account, err := h.multiAcctMgr.GetAccount(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "not_found_error",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account": account,
	})
}

// UpdateAppAccountRequest represents the update request
type UpdateAppAccountRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "active" or "inactive"
}

// UpdateAppAccount updates an app account
// PUT /api/app-accounts/:id
func (h *AppAccountHandler) UpdateAppAccount(c *gin.Context) {
	id := c.Param("id")

	var req UpdateAppAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	if err := h.multiAcctMgr.UpdateAccount(id, req.Name, req.Status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "not_found_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Get updated account
	account, _ := h.multiAcctMgr.GetAccount(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account updated successfully",
		"account": account,
	})
}

// DeleteAppAccount deletes an app account
// DELETE /api/app-accounts/:id
func (h *AppAccountHandler) DeleteAppAccount(c *gin.Context) {
	id := c.Param("id")

	if err := h.multiAcctMgr.DeleteAccount(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"type":    "not_found_error",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account deleted successfully",
	})
}
