package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/tihe/susi-property-service/handlers"
	"github.com/tihe/susi-property-service/models"
	"github.com/tihe/susi-property-service/services"
	"github.com/tihe/susi-shared/eureka"
	"github.com/tihe/susi-shared/events"
)

func main() {
	// Get environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "susi"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	// Database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Kafka producer using shared module
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	kafkaConfig := events.KafkaConfig{
		Brokers: []string{kafkaBrokers},
		Topic:   "property-events",
	}
	producer := events.NewKafkaProducer(kafkaConfig)
	defer producer.Close()

	// Eureka client
	eurekaServerURL := os.Getenv("EUREKA_SERVER_URL")
	if eurekaServerURL == "" {
		eurekaServerURL = "http://localhost:8761/eureka/"
	}
	eurekaClient := eureka.NewEurekaClient(eurekaServerURL)

	// Service configuration
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "property-service"
	}
	servicePortStr := os.Getenv("SERVICE_PORT")
	if servicePortStr == "" {
		servicePortStr = "8082"
	}
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		log.Fatal("Invalid service port:", err)
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

	// Health check endpoint for Eureka
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// API versioning
	apiV1 := router.Group("/api/v1")

	// Property routes
	propertyHandler := handlers.NewPropertyHandler(propertyService)
	handlers.RegisterPropertyRoutes(apiV1, propertyHandler)

	// Room routes
	roomHandler := handlers.NewRoomHandler(roomService)
	handlers.RegisterRoomRoutes(apiV1, roomHandler)

	// LandLord routes
	landLordHandler := handlers.NewLandLordHandler(landLordService)
	handlers.RegisterLandLordRoutes(apiV1, landLordHandler)

	// Graceful shutdown setup
	srv := &http.Server{
		Addr:    ":" + servicePortStr,
		Handler: router,
	}

	// Register with Eureka
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	if err := eurekaClient.Register(serviceName, hostname, servicePort); err != nil {
		log.Printf("Failed to register with Eureka: %v", err)
	} else {
		log.Printf("Successfully registered with Eureka as %s", serviceName)
	}

	// Start heartbeat goroutine
	instanceID := fmt.Sprintf("%s:%s:%d", hostname, serviceName, servicePort)
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	go func() {
		for range heartbeatTicker.C {
			if err := eurekaClient.Heartbeat(serviceName, instanceID); err != nil {
				log.Printf("Failed to send heartbeat to Eureka: %v", err)
			}
		}
	}()

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
