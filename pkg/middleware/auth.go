package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
