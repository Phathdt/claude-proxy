package handlers

import (
	"net/http"

	"claude-proxy/pkg/token"

	"github.com/gin-gonic/gin"
)

// TokensHandler handles token management endpoints
type TokensHandler struct {
	tokenManager *token.Manager
}

// NewTokensHandler creates a new tokens handler
func NewTokensHandler(tokenManager *token.Manager) *TokensHandler {
	return &TokensHandler{
		tokenManager: tokenManager,
	}
}

// CreateTokenRequest represents the request to create a token
type CreateTokenRequest struct {
	Name   string `json:"name" binding:"required"`
	Key    string `json:"key" binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// UpdateTokenRequest represents the request to update a token
type UpdateTokenRequest struct {
	Name   string `json:"name" binding:"required"`
	Key    string `json:"key" binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// ListTokens lists all tokens
// GET /api/tokens
func (h *TokensHandler) ListTokens(c *gin.Context) {
	tokens := h.tokenManager.ListTokens()
	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
	})
}

// GetToken gets a single token
// GET /api/tokens/:id
func (h *TokensHandler) GetToken(c *gin.Context) {
	id := c.Param("id")

	token, err := h.tokenManager.GetToken(id)
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
		"token": token,
	})
}

// CreateToken creates a new token
// POST /api/tokens
func (h *TokensHandler) CreateToken(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	token, err := h.tokenManager.CreateToken(req.Name, req.Key, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Token created successfully",
		"token":   token,
	})
}

// UpdateToken updates a token
// PUT /api/tokens/:id
func (h *TokensHandler) UpdateToken(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	token, err := h.tokenManager.UpdateToken(id, req.Name, req.Key, req.Status)
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
		"success": true,
		"message": "Token updated successfully",
		"token":   token,
	})
}

// DeleteToken deletes a token
// DELETE /api/tokens/:id
func (h *TokensHandler) DeleteToken(c *gin.Context) {
	id := c.Param("id")

	if err := h.tokenManager.DeleteToken(id); err != nil {
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
		"message": "Token deleted successfully",
	})
}

// GenerateKey generates a random API key
// POST /api/tokens/generate-key
func (h *TokensHandler) GenerateKey(c *gin.Context) {
	key, err := token.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "internal_error",
				"message": "Failed to generate key",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key": key,
	})
}
