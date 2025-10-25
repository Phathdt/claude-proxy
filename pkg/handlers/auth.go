package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"claude-proxy/config"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	cfg *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		cfg: cfg,
	}
}

// ValidateRequest represents the API key validation request
type ValidateRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// ValidateResponse represents the validation response
type ValidateResponse struct {
	Valid bool   `json:"valid"`
	User  *User  `json:"user,omitempty"`
	Error string `json:"error,omitempty"`
}

// User represents the authenticated user info
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// Validate validates an API key
// POST /api/auth/validate
func (h *AuthHandler) Validate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ValidateResponse{
			Valid: false,
			Error: "Invalid request body",
		})
		return
	}

	// Validate API key against configured key
	if req.APIKey != h.cfg.Auth.APIKey {
		c.JSON(http.StatusUnauthorized, ValidateResponse{
			Valid: false,
			Error: "Invalid API key",
		})
		return
	}

	// Return success with user info
	c.JSON(http.StatusOK, ValidateResponse{
		Valid: true,
		User: &User{
			ID:    "admin",
			Email: "admin@clove.local",
			Name:  "Clove Admin",
			Role:  "admin",
		},
	})
}

// LoginRequest represents the login request (for UI convenience)
type LoginRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"` // Returns the API key as "token" for frontend
	User    *User  `json:"user,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Login handles login with API key
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate API key
	if req.APIKey != h.cfg.Auth.APIKey {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Invalid API key",
		})
		return
	}

	// Return success
	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   req.APIKey, // Return the API key as token
		User: &User{
			ID:    "admin",
			Email: "admin@clove.local",
			Name:  "Clove Admin",
			Role:  "admin",
		},
	})
}
