package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/tihe/susi-gateway/middleware"
	"github.com/tihe/susi-proto/auth"
	"go-micro.dev/v5"
	"go-micro.dev/v5/client"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceClient struct {
	client client.Client
}

func NewServiceClient(consulURL string) (*ServiceClient, error) {
	reg := consul.NewConsulRegistry(registry.Addrs(consulURL))

	service := micro.NewService(
		micro.Registry(reg),
		micro.Name("api-gateway"),
	)

	return &ServiceClient{
		client: service.Client(),
	}, nil
}

type GenericRequest struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body,omitempty"`
	Query   map[string]string      `json:"query,omitempty"`
}

func proxyRequestToGRPC(c *gin.Context, serviceClient *ServiceClient, serviceName, endpoint string) {

}

func handleAuthService(serviceClient *ServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func proxyRequest(c *gin.Context, serviceDiscovery registry.Registry) {
	// Copy headers

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

	registry := consul.NewConsulRegistry(registry.Addrs(consulURL))

	// TODO: wrap a func to init all client
	authServiceName := os.Getenv("AUTH_SERVICE_NAME")
	if authServiceName == "" {
		authServiceName = "auth-service"
	}

	ctx := context.Background()
	gatewayMux := runtime.NewServeMux()

	authService, err := registry.GetService(authServiceName)
	if err != nil || len(authService) == 0 {
		log.Panicf("Cannot discover auth service: %v", err)
	}
	authConn, err := grpc.NewClient(authService[0].Nodes[0].Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("Cannot connect to auth service: %v", err)
	}
	defer authConn.Close()

	if err := auth.RegisterAuthServiceHandler(ctx, gatewayMux, authConn); err != nil {
		log.Panicf("Cannot register auth service: %v", err)
	}

	// Public routes (no JWT validation required)
	// Auth routes
	r.Any("/api/v1/auth/*proxyPath", func(c *gin.Context) {
		gin.WrapH(gatewayMux)
	})

	// Protected routes (JWT validation required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(registry))

	// // Property routes
	// protected.Any("/properties/*proxyPath", func(c *gin.Context) {
	// 	proxyRequest(c, serviceDiscovery, "property-service")
	// })
	// // Room routes
	// protected.Any("/rooms/*proxyPath", func(c *gin.Context) {
	// 	proxyRequest(c, serviceDiscovery, "property-service")
	// })
	// // Landlord routes
	// protected.Any("/landlords/*proxyPath", func(c *gin.Context) {
	// 	proxyRequest(c, serviceDiscovery, "property-service")
	// })
	// // Tenant routes
	// protected.Any("/tenants/*proxyPath", func(c *gin.Context) {
	// 	proxyRequest(c, serviceDiscovery, "tenant-service")
	// })
	// // Renovation routes
	// protected.Any("/renovations/*proxyPath", func(c *gin.Context) {
	// 	proxyRequest(c, serviceDiscovery, "renovation-service")
	// })

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
