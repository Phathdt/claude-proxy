package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"claude-proxy/config"
	"claude-proxy/pkg/account"
	"claude-proxy/pkg/claude"
	"claude-proxy/pkg/retry"
)

// MessagesHandler handles message-related endpoints
type MessagesHandler struct {
	claudeClient   *claude.Client
	accountManager *account.Manager
	retryConfig    retry.Config
}

// NewMessagesHandler creates a new messages handler
func NewMessagesHandler(claudeClient *claude.Client, accountManager *account.Manager, cfg *config.Config) *MessagesHandler {
	return &MessagesHandler{
		claudeClient:   claudeClient,
		accountManager: accountManager,
		retryConfig: retry.Config{
			MaxRetries: cfg.Retry.MaxRetries,
			RetryDelay: cfg.Retry.RetryDelay,
		},
	}
}

// CreateMessage handles message creation
// POST /v1/messages
func (h *MessagesHandler) CreateMessage(c *gin.Context) {
	var req claude.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": fmt.Sprintf("Invalid request body: %v", err),
			},
		})
		return
	}

	// Set default model if not specified
	if req.Model == "" {
		req.Model = "claude-opus-4-20250514"
	}

	// Set default max_tokens if not specified
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get valid access token (will refresh if needed)
	accessToken, err := h.accountManager.GetValidToken(ctx)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"type":    "authentication_error",
				"message": fmt.Sprintf("Failed to get valid token: %v", err),
			},
		})
		return
	}

	// Handle streaming vs non-streaming
	if req.Stream {
		h.handleStreamingRequest(c, ctx, accessToken, &req)
	} else {
		h.handleNonStreamingRequest(c, ctx, accessToken, &req)
	}
}

// handleNonStreamingRequest handles non-streaming message requests
func (h *MessagesHandler) handleNonStreamingRequest(c *gin.Context, ctx context.Context, accessToken string, req *claude.MessageRequest) {
	// Send message with retry logic
	resp, err := retry.DoWithResult(ctx, h.retryConfig, func() (*claude.MessageResponse, error) {
		return h.claudeClient.SendMessage(ctx, accessToken, req)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "api_error",
				"message": fmt.Sprintf("Failed to send message: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// handleStreamingRequest handles streaming message requests
func (h *MessagesHandler) handleStreamingRequest(c *gin.Context, ctx context.Context, accessToken string, req *claude.MessageRequest) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Get event channel from Claude client
	eventChan, errChan := h.claudeClient.SendMessageStream(ctx, accessToken, req)

	// Use Gin's built-in streaming support
	c.Writer.WriteHeader(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "streaming_error",
				"message": "Streaming not supported",
			},
		})
		return
	}

	// Stream events to client
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				// Channel closed, streaming complete
				return
			}

			// Write SSE event
			if event.Event != "" {
				fmt.Fprintf(c.Writer, "event: %s\n", event.Event)
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", event.Data)
			flusher.Flush()

		case err := <-errChan:
			if err != nil {
				// Send error event
				errorData := map[string]interface{}{
					"type":  "error",
					"error": map[string]interface{}{
						"type":    "api_error",
						"message": err.Error(),
					},
				}
				data, _ := json.Marshal(errorData)
				fmt.Fprintf(c.Writer, "event: error\n")
				fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
				flusher.Flush()
			}
			return

		case <-ctx.Done():
			// Context cancelled
			return
		}
	}
}
