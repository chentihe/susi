package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tihe/susi-gateway/gateway"
	"github.com/tihe/susi-gateway/middleware"
	"github.com/tihe/susi-shared/discovery/consul"
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

	ctx := context.Background()
	gw, err := gateway.NewGateway(ctx, registry)
	if err != nil {
		log.Fatal("Failed to initialize gateway")
	}

	// Public routes (no JWT validation required)
	r.Use(middleware.JWTAuthMiddleware(registry))
	r.Any("/api/v1/*proxyPath", gin.WrapH(gw))

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
