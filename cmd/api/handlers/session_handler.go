package handlers

import (
	"net/http"

	"claude-proxy/modules/proxy/application/dto"
	"claude-proxy/modules/proxy/domain/interfaces"
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
			AccountID:   session.AccountID,
			TokenID:     session.TokenID,
			UserAgent:   session.UserAgent,
			IPAddress:   session.IPAddress,
			CreatedAt:   session.CreatedAt.Unix(),
			LastSeenAt:  session.LastSeenAt.Unix(),
			ExpiresAt:   session.ExpiresAt.Unix(),
			IsActive:    session.IsActive,
			RequestPath: session.RequestPath,
		}
	}

	c.JSON(http.StatusOK, dto.ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	})
}

// ListAccountSessions lists sessions for a specific account
// GET /api/accounts/:id/sessions
func (h *SessionHandler) ListAccountSessions(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("id")

	sessions, err := h.sessionService.GetAccountSessions(ctx, accountID)
	if err != nil {
		h.logger.Withs(sctx.Fields{"error": err, "account_id": accountID}).Error("Failed to list account sessions")
		panic(errors.NewInternalServerError("failed to list account sessions: " + err.Error()))
	}

	// Convert to DTOs
	sessionResponses := make([]*dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = &dto.SessionResponse{
			ID:          session.ID,
			AccountID:   session.AccountID,
			TokenID:     session.TokenID,
			UserAgent:   session.UserAgent,
			IPAddress:   session.IPAddress,
			CreatedAt:   session.CreatedAt.Unix(),
			LastSeenAt:  session.LastSeenAt.Unix(),
			ExpiresAt:   session.ExpiresAt.Unix(),
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

// RevokeAccountSessions revokes all sessions for an account
// DELETE /api/accounts/:id/sessions
func (h *SessionHandler) RevokeAccountSessions(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("id")

	count, err := h.sessionService.RevokeAccountSessions(ctx, accountID)
	if err != nil {
		h.logger.Withs(sctx.Fields{"error": err, "account_id": accountID}).Error("Failed to revoke account sessions")
		panic(errors.NewInternalServerError("failed to revoke account sessions: " + err.Error()))
	}

	h.logger.Withs(sctx.Fields{
		"account_id":    accountID,
		"revoked_count": count,
	}).Info("Account sessions revoked via API")

	c.JSON(http.StatusOK, dto.RevokeAccountSessionsResponse{
		Success:      true,
		Message:      "Account sessions revoked successfully",
		RevokedCount: count,
	})
}
