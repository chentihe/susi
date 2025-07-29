package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=susi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Kafka producer
	producer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "auth-events",
		Balancer: &kafka.LeastBytes{},
	}

	// Initialize repositories
	adminRepo := models.NewAdminImpl(db)

	// Initialize services
	adminService := services.NewAdminService(adminRepo, producer)

	// Setup router
	router := gin.Default()
	
	// API versioning
	apiV1 := router.Group("/api/v1")
	
	// Auth routes (public)
	authHandler := handlers.NewAuthHandler(adminService)
	authHandler.RegisterAuthRoutes(apiV1)

	// Graceful shutdown setup
	srv := &http.Server{
		Addr:    ":8081", // Auth service port
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down auth service...")

	// The context is used to inform the server it has 5 seconds to finish
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Auth service exiting")
} 