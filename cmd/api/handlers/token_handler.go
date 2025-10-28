package handlers

import (
	"net/http"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/auth/domain/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/phathdt/service-context/core"
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

// ListTokens lists all tokens with optional filtering and pagination
// GET /api/tokens?role=admin&status=active&search=prod&page=1&limit=10
func (h *TokenHandler) ListTokens(c *gin.Context) {
	// Parse query parameters
	var query dto.TokenQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Parse pagination parameters
	var paging core.Paging
	if err := c.ShouldBindQuery(&paging); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Process pagination params (normalize page/limit)
	paging.Process()

	// Get tokens from service (paging is mutated with metadata)
	tokens, err := h.tokenService.ListTokens(c.Request.Context(), &query, &paging)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "internal_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Build response with paging metadata
	c.JSON(http.StatusOK, gin.H{
		"tokens": dto.ToTokenResponses(tokens),
		"paging": paging,
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
		entities.TokenRole(req.Role),
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
		entities.TokenRole(req.Role),
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
