package middleware

import (
	"net/http"
	"strings"

	"claude-proxy/pkg/errors"
	"claude-proxy/pkg/token"

	"github.com/gin-gonic/gin"
	sctx "github.com/phathdt/service-context"
)

// APIKeyAuth creates middleware for API key authentication
func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		providedKey := c.GetHeader("X-API-Key")

		// Check if API key is provided and matches
		if providedKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"type":    "authentication_error",
					"message": "API key is required",
				},
			})
			c.Abort()
			return
		}

		if providedKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"type":    "authentication_error",
					"message": "Invalid API key",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// BearerTokenAuth creates middleware for Bearer token authentication
func BearerTokenAuth(tokenManager *token.Manager, logger sctx.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			panic(errors.NewUnauthorizedError("missing authorization header"))
		}

		// Parse bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			panic(errors.NewUnauthorizedError("invalid authorization header format, expected 'Bearer <token>'"))
		}
		bearerToken := parts[1]

		// Validate token using token manager
		validatedToken, err := tokenManager.ValidateToken(bearerToken)
		if err != nil {
			logger.Withs(sctx.Fields{
				"error": err.Error(),
			}).Warn("Token validation failed")
			panic(errors.NewUnauthorizedError("invalid or inactive token"))
		}

		logger.Withs(sctx.Fields{
			"token_id":   validatedToken.ID,
			"token_name": validatedToken.Name,
			"path":       c.Request.URL.Path,
		}).Info("Token validated successfully")

		// Store validated token in context for handler use
		c.Set("validated_token", validatedToken)
		c.Next()
	}
}

