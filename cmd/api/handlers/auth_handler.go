package handlers

import (
	"net/http"

	"claude-proxy/config"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	apiKey string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		apiKey: cfg.Auth.APIKey,
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

	// Validate API key
	if req.APIKey != h.apiKey {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API key",
		})
		return
	}

	// Return success with token (just return the API key as token)
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

	// Validate API key
	if req.APIKey != h.apiKey {
		c.JSON(http.StatusOK, ValidateResponse{
			Valid: false,
		})
		return
	}

	// Return valid with user info
	c.JSON(http.StatusOK, ValidateResponse{
		Valid: true,
		User: &User{
			ID:    "admin",
			Email: "admin@localhost",
			Name:  "Admin",
			Role:  "admin",
		},
	})
}
