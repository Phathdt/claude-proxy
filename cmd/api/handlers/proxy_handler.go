package handlers

import (
	"context"
	"io"

	"claude-proxy/modules/auth/domain/entities"
	"claude-proxy/modules/proxy/domain/interfaces"
	"claude-proxy/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ProxyHandler handles HTTP requests for proxying to Claude API
type ProxyHandler struct {
	proxyService interfaces.ProxyService
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(proxyService interfaces.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
	}
}

// ProxyRequest proxies a request to Claude API
func (h *ProxyHandler) ProxyRequest(c *gin.Context) {
	// Get validated token from context (set by BearerTokenAuth middleware)
	validatedToken, exists := c.Get("validated_token")
	if !exists {
		panic(errors.NewUnauthorizedError("token not found in context"))
	}
	userToken := validatedToken.(*entities.Token)

	// Proxy the request
	resp, err := h.proxyService.ProxyRequest(c.Request.Context(), userToken, c.Request)
	if err != nil {
		// Check if context was canceled or timed out
		ctxErr := c.Request.Context().Err()
		if err == context.Canceled || ctxErr == context.Canceled {
			// Don't panic for canceled requests - just abort silently
			c.AbortWithStatus(499) // 499 Client Closed Request (nginx convention)
			return
		}
		if err == context.DeadlineExceeded || ctxErr == context.DeadlineExceeded {
			// Request timed out
			panic(errors.NewRequestTimeoutError("request timed out"))
		}
		panic(errors.NewServiceUnavailableError(err.Error()))
	}
	defer resp.Body.Close()

	// Copy response headers first (before streaming or buffering)
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Check if response is SSE (Server-Sent Events) stream
	contentType := resp.Header.Get("Content-Type")
	if contentType == "text/event-stream" {
		// Stream SSE response directly to client
		h.streamSSEResponse(c, &resp.Body)
		return
	}

	// For non-streaming responses, buffer and send (existing behavior)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// Check if context was canceled while reading response
		if c.Request.Context().Err() == context.Canceled {
			c.AbortWithStatus(499) // 499 Client Closed Request
			return
		}
		panic(errors.NewInternalServerError("failed to read response body"))
	}

	// Return buffered response
	c.Data(resp.StatusCode, contentType, respBody)
}

// streamSSEResponse streams Server-Sent Events from Claude API to the client using Gin's Stream
func (h *ProxyHandler) streamSSEResponse(c *gin.Context, resp *io.ReadCloser) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Use Gin's Stream method for efficient streaming
	c.Stream(func(w io.Writer) bool {
		// Check if context was canceled
		select {
		case <-c.Request.Context().Done():
			// Client disconnected or timeout - stop streaming
			return false
		default:
			// Continue streaming
		}

		// Read and write chunks from Claude API to client
		buf := make([]byte, 4096) // 4KB buffer for streaming
		n, err := (*resp).Read(buf)

		if n > 0 {
			// Write chunk to client
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				// Client disconnected or write error - stop streaming
				return false
			}
		}

		// Check for errors
		if err == io.EOF {
			// End of stream - stop streaming
			return false
		}
		if err != nil {
			// Stream error - stop streaming
			return false
		}

		// Continue streaming (return true to keep stream open)
		return true
	})
}

// GetModels handles GET /v1/models
func (h *ProxyHandler) GetModels(c *gin.Context) {
	h.ProxyRequest(c)
}

// CreateMessage handles POST /v1/messages
func (h *ProxyHandler) CreateMessage(c *gin.Context) {
	h.ProxyRequest(c)
}
