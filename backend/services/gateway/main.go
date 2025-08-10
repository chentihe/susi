package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tihe/susi-gateway/middleware"
	"github.com/tihe/susi-shared/eureka"
)

type ServiceDiscovery struct {
	eurekaClient *eureka.EurekaClient
}

func NewServiceDiscovery(eurekaServerURL string) *ServiceDiscovery {
	return &ServiceDiscovery{
		eurekaClient: eureka.NewEurekaClient(eurekaServerURL),
	}
}

func (sd *ServiceDiscovery) GetServiceURL(serviceName string) (string, error) {
	// Use health check when getting service URL
	return sd.eurekaClient.GetServiceURLWithHealthCheck(serviceName)
}

func proxyRequest(c *gin.Context, serviceDiscovery *ServiceDiscovery, serviceName string) {
	// Get service URL from Eureka
	targetBase, err := serviceDiscovery.GetServiceURL(serviceName)
	if err != nil {
		log.Printf("Failed to get service URL for %s: %v", serviceName, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service unavailable"})
		return
	}

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
		log.Printf("Failed to proxy request to %s: %v", serviceName, err)
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

	eurekaServerURL := os.Getenv("EUREKA_SERVER_URL")
	if eurekaServerURL == "" {
		eurekaServerURL = "http://localhost:8761/eureka/"
	}

	serviceDiscovery := NewServiceDiscovery(eurekaServerURL)

	// Public routes (no JWT validation required)
	// Auth routes
	r.Any("/api/v1/auth/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, serviceDiscovery, "auth-service")
	})

	// Protected routes (JWT validation required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(eurekaServerURL))

	// Property routes
	protected.Any("/properties/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, serviceDiscovery, "property-service")
	})
	// Room routes
	protected.Any("/rooms/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, serviceDiscovery, "property-service")
	})
	// Landlord routes
	protected.Any("/landlords/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, serviceDiscovery, "property-service")
	})
	// Tenant routes
	protected.Any("/tenants/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, serviceDiscovery, "tenant-service")
	})
	// Renovation routes
	protected.Any("/renovations/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, serviceDiscovery, "renovation-service")
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
