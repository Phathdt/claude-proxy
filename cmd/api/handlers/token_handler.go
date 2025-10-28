package handlers

import (
	"net/http"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	"github.com/gin-gonic/gin"
)

// TokenHandler handles HTTP requests for token management
type TokenHandler struct {
	tokenService interfaces.TokenService
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(tokenService interfaces.TokenService) *TokenHandler {
	return &TokenHandler{
		tokenService: tokenService,
	}
}

// ListTokens lists all tokens
// GET /api/tokens
func (h *TokenHandler) ListTokens(c *gin.Context) {
	tokens, err := h.tokenService.ListTokens(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "internal_error",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": dto.ToTokenResponses(tokens),
	})
}

// GetToken gets a single token
// GET /api/tokens/:id
func (h *TokenHandler) GetToken(c *gin.Context) {
	id := c.Param("id")

	token, err := h.tokenService.GetTokenByID(c.Request.Context(), id)
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
		"token": dto.ToTokenResponse(token),
	})
}

// CreateToken creates a new token
// POST /api/tokens
func (h *TokenHandler) CreateToken(c *gin.Context) {
	var req dto.CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Call service to create token
	token, err := h.tokenService.CreateToken(
		c.Request.Context(),
		req.Name,
		req.Key,
		entities.TokenStatus(req.Status),
	)
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
		"token":   dto.ToTokenResponse(token),
	})
}

// UpdateToken updates a token
// PUT /api/tokens/:id
func (h *TokenHandler) UpdateToken(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Call service to update token
	token, err := h.tokenService.UpdateToken(
		c.Request.Context(),
		id,
		req.Name,
		req.Key,
		entities.TokenStatus(req.Status),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token updated successfully",
		"token":   dto.ToTokenResponse(token),
	})
}

// DeleteToken deletes a token
// DELETE /api/tokens/:id
func (h *TokenHandler) DeleteToken(c *gin.Context) {
	id := c.Param("id")

	if err := h.tokenService.DeleteToken(c.Request.Context(), id); err != nil {
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
