package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// OpenAICompatibility middleware transforms OpenAI-format requests to Claude format
func OpenAICompatibility() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only transform POST requests to /v1/chat/completions
		if c.Request.Method == "POST" && strings.HasSuffix(c.Request.URL.Path, "/v1/chat/completions") {
			// Read request body
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Next()
				return
			}

			// Parse OpenAI request
			var openAIReq map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &openAIReq); err != nil {
				// If parsing fails, restore body and continue
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				c.Next()
				return
			}

			// Transform to Claude format
			claudeReq := make(map[string]interface{})

			// Copy model
			if model, ok := openAIReq["model"].(string); ok {
				claudeReq["model"] = model
			}

			// Copy messages
			if messages, ok := openAIReq["messages"].([]interface{}); ok {
				claudeReq["messages"] = messages
			}

			// Copy max_tokens (required by Claude)
			if maxTokens, ok := openAIReq["max_tokens"].(float64); ok {
				claudeReq["max_tokens"] = int(maxTokens)
			} else {
				// Default to 1024 if not specified
				claudeReq["max_tokens"] = 1024
			}

			// Copy temperature
			if temp, ok := openAIReq["temperature"].(float64); ok {
				claudeReq["temperature"] = temp
			}

			// Copy stream
			if stream, ok := openAIReq["stream"].(bool); ok {
				claudeReq["stream"] = stream
			}

			// Convert back to JSON
			newBody, err := json.Marshal(claudeReq)
			if err != nil {
				// If marshaling fails, restore original body
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				c.Next()
				return
			}

			// Replace request body
			c.Request.Body = io.NopCloser(bytes.NewReader(newBody))
			c.Request.ContentLength = int64(len(newBody))

			// Rewrite path to /v1/messages
			c.Request.URL.Path = "/v1/messages"
		}

		c.Next()
	}
}
