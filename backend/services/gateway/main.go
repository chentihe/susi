package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tihe/susi-gateway/middleware"
)

func proxyRequest(c *gin.Context, targetBase string) {
	// Build the target URL
	targetURL := targetBase + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Create the new request
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}
	// Copy headers
	for k, v := range c.Request.Header {
		req.Header[k] = v
	}

	// Do the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers and status
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func main() {
	r := gin.Default()

	authURL := os.Getenv("AUTH_SERVICE_URL")
	propertyURL := os.Getenv("PROPERTY_SERVICE_URL")
	tenantURL := os.Getenv("TENANT_SERVICE_URL")
	renovationURL := os.Getenv("RENOVATION_SERVICE_URL")

	// Public routes (no JWT validation required)
	// Auth routes
	r.Any("/api/v1/auth/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, authURL)
	})

	// Protected routes (JWT validation required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(authURL))

	// Property routes
	protected.Any("/properties/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, propertyURL)
	})
	// Room routes
	protected.Any("/rooms/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, propertyURL)
	})
	// Landlord routes
	protected.Any("/landlords/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, propertyURL)
	})
	// Tenant routes
	protected.Any("/tenants/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, tenantURL)
	})
	// Renovation routes
	protected.Any("/renovations/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, renovationURL)
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway listening on :%s", port)
	r.Run(":" + port)
}
