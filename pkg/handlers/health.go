package handlers

import (
	"net/http"
	"time"

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
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"account":   status,
	})
}
