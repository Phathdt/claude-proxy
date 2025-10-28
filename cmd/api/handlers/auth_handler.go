package handlers

import (
	"net/http"

	"claude-proxy/config"
	"claude-proxy/modules/auth/domain/interfaces"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	tokenService interfaces.TokenService
	configAPIKey string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(tokenService interfaces.TokenService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		tokenService: tokenService,
		configAPIKey: cfg.Auth.APIKey,
	}
}

// LoginRequest is the request body for login
type LoginRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// LoginResponse is the response body for login
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	User    User   `json:"user"`
}

// User represents a user in the system
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// ValidateRequest is the request body for validate
type ValidateRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// ValidateResponse is the response body for validate
type ValidateResponse struct {
	Valid bool  `json:"valid"`
	User  *User `json:"user,omitempty"`
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// First, try to validate as admin token from token service
	token, err := h.tokenService.ValidateToken(c.Request.Context(), req.APIKey)
	if err == nil && token.IsAdmin() {
		// Valid admin token
		c.JSON(http.StatusOK, LoginResponse{
			Success: true,
			Token:   req.APIKey,
			User: User{
				ID:    token.ID,
				Email: "admin@localhost",
				Name:  token.Name,
				Role:  string(token.Role),
			},
		})
		return
	}

	// Fall back to config API key (backward compatibility)
	if req.APIKey == h.configAPIKey {
		c.JSON(http.StatusOK, LoginResponse{
			Success: true,
			Token:   req.APIKey,
			User: User{
				ID:    "admin",
				Email: "admin@localhost",
				Name:  "Admin",
				Role:  "admin",
			},
		})
		return
	}

	// Neither admin token nor config key
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Invalid API key",
	})
}

// Validate handles POST /api/auth/validate
func (h *AuthHandler) Validate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// First, try to validate as admin token from token service
	token, err := h.tokenService.ValidateToken(c.Request.Context(), req.APIKey)
	if err == nil && token.IsAdmin() {
		// Valid admin token
		c.JSON(http.StatusOK, ValidateResponse{
			Valid: true,
			User: &User{
				ID:    token.ID,
				Email: "admin@localhost",
				Name:  token.Name,
				Role:  string(token.Role),
			},
		})
		return
	}

	// Fall back to config API key (backward compatibility)
	if req.APIKey == h.configAPIKey {
		c.JSON(http.StatusOK, ValidateResponse{
			Valid: true,
			User: &User{
				ID:    "admin",
				Email: "admin@localhost",
				Name:  "Admin",
				Role:  "admin",
			},
		})
		return
	}

	// Neither admin token nor config key
	c.JSON(http.StatusOK, ValidateResponse{
		Valid: false,
	})
}
