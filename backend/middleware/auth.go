package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/arandu-ai/arandu/config"
	"github.com/gin-gonic/gin"
)

const (
	// APIKeyHeader is the header name for API key authentication
	APIKeyHeader = "X-API-Key"

	// AuthorizationHeader is the standard Authorization header
	AuthorizationHeader = "Authorization"
)

// APIKeyAuth returns a middleware that validates API key authentication
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if API key authentication is not required
		if !config.Config.RequireAPIKey || config.Config.APIKey == "" {
			c.Next()
			return
		}

		// Skip authentication for health check endpoint
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}

		// Try X-API-Key header first
		apiKey := c.GetHeader(APIKeyHeader)

		// Fall back to Authorization header (Bearer token style)
		if apiKey == "" {
			authHeader := c.GetHeader(AuthorizationHeader)
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				apiKey = authHeader[7:]
			}
		}

		// Validate API key using constant-time comparison
		if !secureCompare(apiKey, config.Config.APIKey) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing API key",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		c.Next()
	}
}

// secureCompare performs a constant-time string comparison
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// OptionalAPIKeyAuth returns a middleware that checks API key if provided
// but doesn't require it
func OptionalAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Config.APIKey == "" {
			c.Next()
			return
		}

		apiKey := c.GetHeader(APIKeyHeader)
		if apiKey == "" {
			authHeader := c.GetHeader(AuthorizationHeader)
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				apiKey = authHeader[7:]
			}
		}

		// If no API key provided, continue without authentication
		if apiKey == "" {
			c.Next()
			return
		}

		// If API key is provided, validate it
		if !secureCompare(apiKey, config.Config.APIKey) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Mark request as authenticated
		c.Set("authenticated", true)
		c.Next()
	}
}

// RequireAuth returns a middleware that requires authentication
// Use this for sensitive endpoints
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Config.RequireAPIKey {
			c.Next()
			return
		}

		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		c.Next()
	}
}
