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

	"github.com/tihe/susi-auth-service/handlers"
	"github.com/tihe/susi-auth-service/models"
	"github.com/tihe/susi-auth-service/services"
	"github.com/tihe/susi-shared/discovery/consul"
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
		Topic:   "auth-events",
	}
	producer := events.NewKafkaProducer(kafkaConfig)
	defer producer.Close()

	// Consul registration
	consulURL := os.Getenv("CONSUL_SERVER_URL")
	consulClient, err := consul.NewConsulClient(consulURL)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	// Service configuration
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "auth-service"
	}
	servicePortStr := os.Getenv("SERVICE_PORT")
	if servicePortStr == "" {
		servicePortStr = "8081"
	}
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		log.Fatal("Invalid service port:", err)
	}

	// Initialize JWT key
	if err := services.InitJWTKey(); err != nil {
		log.Fatal("Failed to initialize JWT key:", err)
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
	authHandler := handlers.NewAuthHandler(db, adminService)
	handlers.RegisterAuthRoutes(apiV1, authHandler)

	// Graceful shutdown setup
	srv := &http.Server{
		Addr:    ":" + servicePortStr,
		Handler: router,
	}

	// Register with Consul
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	serviceId := fmt.Sprintf("%s:%s:%d", serviceName, hostname, servicePort)

	if err := consulClient.Register(serviceName, hostname, servicePort); err != nil {
		log.Printf("Failed to register with Consul: %v", err)
	} else {
		log.Printf("Successfully registered with Consul as %s", serviceName)
	}

	// Start heartbeat goroutine
	go func() {
		for {
			if err := consulClient.HealthCheck(serviceId); err != nil {
				log.Println("Failed to report healthy state: ", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer func() {
		if err := consulClient.Deregister(serviceName, hostname, servicePort); err != nil {
			log.Printf("Failed to deregister from Consul: %v", err)
		} else {
			log.Println("Successfully deregistered from Consul")
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
	log.Println("Shutting down auth service...")

	// The context is used to inform the server it has 5 seconds to finish
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Auth service exiting")
}
