package handlers

import (
	"net/http"

	"claude-proxy/modules/auth/application/dto"
	"claude-proxy/modules/auth/domain/interfaces"
	"claude-proxy/pkg/errors"

	"github.com/gin-gonic/gin"
	sctx "github.com/phathdt/service-context"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	sessionService interfaces.SessionService
	logger         sctx.Logger
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	sessionService interfaces.SessionService,
	appLogger sctx.Logger,
) *SessionHandler {
	logger := appLogger.Withs(sctx.Fields{"component": "session-handler"})
	return &SessionHandler{
		sessionService: sessionService,
		logger:         logger,
	}
}

// ListAllSessions lists all active sessions (admin)
// GET /api/admin/sessions
func (h *SessionHandler) ListAllSessions(c *gin.Context) {
	ctx := c.Request.Context()

	sessions, err := h.sessionService.GetAllSessions(ctx)
	if err != nil {
		h.logger.Withs(sctx.Fields{"error": err}).Error("Failed to list sessions")
		panic(errors.NewInternalServerError("failed to list sessions: " + err.Error()))
	}

	// Convert to DTOs
	sessionResponses := make([]*dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = &dto.SessionResponse{
			ID:          session.ID,
			TokenID:     session.TokenID,
			UserAgent:   session.UserAgent,
			IPAddress:   session.IPAddress,
			CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), // RFC3339
			LastSeenAt:  session.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"), // RFC3339
			ExpiresAt:   session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),  // RFC3339
			IsActive:    session.IsActive,
			RequestPath: session.RequestPath,
		}
	}

	c.JSON(http.StatusOK, dto.ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	})
}

// RevokeSession revokes a specific session
// DELETE /api/sessions/:id
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID := c.Param("id")

	if err := h.sessionService.RevokeSession(ctx, sessionID); err != nil {
		h.logger.Withs(sctx.Fields{"error": err, "session_id": sessionID}).Error("Failed to revoke session")
		panic(errors.NewInternalServerError("failed to revoke session: " + err.Error()))
	}

	h.logger.Withs(sctx.Fields{"session_id": sessionID}).Info("Session revoked via API")

	c.JSON(http.StatusOK, dto.RevokeSessionResponse{
		Success: true,
		Message: "Session revoked successfully",
	})
}
