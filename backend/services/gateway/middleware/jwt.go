package middleware

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware(authServiceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			return
		}

		// Create request to auth service
		req, err := http.NewRequest("POST", authServiceURL+"/api/v1/auth/validate", nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create validation request"})
			return
		}
		req.Header.Set("Authorization", token)

		// Make request to auth service
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "Auth service unavailable"})
			return
		}
		defer resp.Body.Close()

		// Read response body to ensure it's properly closed
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process auth response"})
			return
		}

		// Check if token is valid
		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Token is valid, continue to next handler
		c.Next()
	}
}
