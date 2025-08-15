package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/tihe/susi-auth-service/internal/handler"
	"github.com/tihe/susi-auth-service/internal/repository"
	"github.com/tihe/susi-auth-service/internal/service"
	"github.com/tihe/susi-proto/auth"
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
	registry, err := consul.NewConsulClient(consulURL)
	if err != nil {
		log.Fatal("Failed to initialize registry")
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

	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	serviceId := fmt.Sprintf("%s:%s:%d", serviceName, hostname, servicePort)

	// Initialize JWT key
	if err := service.InitJWTKey(); err != nil {
		log.Fatal("Failed to initialize JWT key:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, producer)

	// Auth routes (public)
	authHanlder := handler.NewAuthHandler(authService)

	grpcServer := grpc.NewServer()
	auth.RegisterAuthServiceServer(grpcServer, authHanlder)

	// Create gRPC listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
	if err != nil {
		log.Printf("failed to listen on gRPC port %d: %v", servicePort, err)
	}

	if err := registry.Register(serviceName, hostname, servicePort); err != nil {
		log.Printf("Failed to register with Consul: %v", err)
	} else {
		log.Printf("Successfully registered with Consul as %s", serviceName)
	}

	// Start heartbeat goroutine
	go func() {
		for {
			if err := registry.HealthCheck(serviceId); err != nil {
				log.Printf("Failed to report healthy state: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Graceful shutdown setup
	go func() {
		log.Printf("gRPC server listening on %s:%d", hostname, servicePort)
		log.Printf("Service: %s", serviceName)

		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down auth service...")

	// The context is used to inform the server it has 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := registry.Deregister(serviceName, hostname, servicePort); err != nil {
		log.Printf("Failed to deregister from Consul: %v", err)
	} else {
		log.Println("Successfully deregistered from Consul")
	}

	// Graceful stop gRPC server
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gRPC server stopped gracefully")
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded, forcing stop")
		grpcServer.Stop()
	}

	log.Println("All servers stopped")
}
