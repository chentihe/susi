package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tihe/susi-gateway/middleware"
	"github.com/tihe/susi-proto/admin"
	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"go-micro.dev/v5/transport/grpc"
)

func proxyRequest(c *gin.Context, serviceDiscovery registry.Registry) {
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

	consulURL := os.Getenv("CONSUL_SERVER_URL")
	if consulURL == "" {
		consulURL = "http://localhost:8500"
	}

	// TODO: wrap a func to init all client
	authServiceName := os.Getenv("AUTH_SERVICE_NAME")
	if authServiceName == "" {
		authServiceName = "auth-service"
	}

	service := micro.NewService(
		micro.Name(authServiceName),
		micro.Registry(consul.NewConsulRegistry(registry.Addrs(consulURL))),
		micro.Transport(grpc.NewTransport()),
		micro.AfterStop(func() error {
			// TODO: add graceful shutdown process
			log.Println("api gateway exiting")
			return nil
		}),
	)

	service.Init()

	adminClient := admin.NewAdminService(authServiceName, service.Client())

	// Public routes (no JWT validation required)
	// Auth routes
	r.Any("/api/v1/auth/*proxyPath", func(c *gin.Context) {
		proxyRequest(c, adminClient)
	})

	// Protected routes (JWT validation required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(adminClient))

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
