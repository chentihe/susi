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
	dsn := "#TODO: using env variable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Kafka producer
	producer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "property-events",
		Balancer: &kafka.LeastBytes{},
	}

	// Initialize repositories
	propertyRepo := models.NewPropertyImpl(db)
	roomRepo := models.NewRoomImpl(db)
	landLordRepo := models.NewLandLordImpl(db)

	// Initialize services
	propertyService := services.NewPropertyService(propertyRepo, producer)
	roomService := services.NewRoomService(roomRepo, producer)
	landLordService := services.NewLandLordService(landLordRepo, producer)

	// Setup router
	router := gin.Default()
	
	// API versioning
	apiV1 := router.Group("/api/v1")
	
	// Protected routes (require JWT)
	protected := apiV1.Group("")
	protected.Use(middleware.JWTAuthMiddleware())
	
	// Property routes
	propertyHandler := handlers.NewPropertyHandler(propertyService)
	propertyHandler.RegisterPropertyRoutes(protected)
	
	// Room routes
	roomHandler := handlers.NewRoomHandler(roomService)
	roomHandler.RegisterRoomRoutes(protected)
	
	// LandLord routes
	landLordHandler := handlers.NewLandLordHandler(landLordService)
	landLordHandler.RegisterLandLordRoutes(protected)

	// Graceful shutdown setup
	srv := &http.Server{
		Addr:    ":8082", // Property service port
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
	log.Println("Shutting down property service...")

	// The context is used to inform the server it has 5 seconds to finish
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Property service exiting")
} 