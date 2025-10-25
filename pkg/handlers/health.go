package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"claude-proxy/pkg/account"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	accountManager *account.Manager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(accountManager *account.Manager) *HealthHandler {
	return &HealthHandler{
		accountManager: accountManager,
	}
}

// Check handles health check requests
// GET /health
func (h *HealthHandler) Check(c *gin.Context) {
	status := h.accountManager.GetStatus()

	// Build response matching MVP spec format
	response := gin.H{
		"status":        "ok",
		"account_valid": status["account_valid"],
	}

	// Add optional fields if account is valid
	if status["account_valid"] == true {
		if expiresAt, ok := status["expires_at"]; ok {
			response["expires_at"] = expiresAt
		}
		if org, ok := status["organization"]; ok {
			response["organization_uuid"] = org
		}
	}

	c.JSON(http.StatusOK, response)
}
