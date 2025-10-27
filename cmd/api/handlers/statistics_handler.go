package handlers

import (
	"net/http"

	"claude-proxy/modules/proxy/domain/interfaces"

	"github.com/gin-gonic/gin"
	sctx "github.com/phathdt/service-context"
)

// StatisticsHandler handles statistics-related requests
type StatisticsHandler struct {
	accountService interfaces.AccountService
	logger         sctx.Logger
}

// NewStatisticsHandler creates a new statistics handler
func NewStatisticsHandler(
	accountService interfaces.AccountService,
	logger sctx.Logger,
) *StatisticsHandler {
	return &StatisticsHandler{
		accountService: accountService,
		logger:         logger,
	}
}

// GetStatistics handles GET /api/admin/statistics
func (h *StatisticsHandler) GetStatistics(c *gin.Context) {
	statistics, err := h.accountService.GetStatistics(c.Request.Context())
	if err != nil {
		h.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to get statistics")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics",
		})
		return
	}

	h.logger.Debug("Statistics retrieved successfully")

	c.JSON(http.StatusOK, statistics)
}
