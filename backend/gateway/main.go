package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/tihe/susi-gateway/middleware"
	"github.com/tihe/susi-proto/auth"
	"github.com/tihe/susi-shared/discovery/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	r := gin.Default()

	consulURL := os.Getenv("CONSUL_SERVER_URL")
	if consulURL == "" {
		consulURL = "http://localhost:8500"
	}

	registry, err := consul.NewConsulClient(consulURL)
	if err != nil {
		log.Fatal("Failed to initialize registry")
	}

	// TODO: wrap a func to init all client
	authServiceName := os.Getenv("AUTH_SERVICE_NAME")
	if authServiceName == "" {
		authServiceName = "auth-service"
	}

	ctx := context.Background()
	gatewayMux := runtime.NewServeMux()

	var authServiceURL string
	for {
		authServiceURL, err = registry.GetServiceURL(authServiceName)
		if err != nil {
			log.Printf("Cannot discover auth service: %v", err)
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("Discover auth service url: %s", authServiceURL)
			break
		}
	}

	authConn, err := grpc.NewClient(authServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("Cannot connect to auth service: %v", err)
	}
	defer authConn.Close()

	if err := auth.RegisterAuthServiceHandler(ctx, gatewayMux, authConn); err != nil {
		log.Panicf("Cannot register auth service: %v", err)
	}

	// Public routes (no JWT validation required)
	// Auth routes
	r.Any("/api/v1/auth/*proxyPath", gin.WrapH(gatewayMux))

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
