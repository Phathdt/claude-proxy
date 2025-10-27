package handlers

import (
	"context"
	"io"

	"claude-proxy/modules/proxy/domain/entities"
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

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// Check if context was canceled while reading response
		if c.Request.Context().Err() == context.Canceled {
			c.AbortWithStatus(499) // 499 Client Closed Request
			return
		}
		panic(errors.NewInternalServerError("failed to read response body"))
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Return response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// GetModels handles GET /v1/models
func (h *ProxyHandler) GetModels(c *gin.Context) {
	h.ProxyRequest(c)
}

// CreateMessage handles POST /v1/messages
func (h *ProxyHandler) CreateMessage(c *gin.Context) {
	h.ProxyRequest(c)
}
