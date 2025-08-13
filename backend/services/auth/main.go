package main

import (
	"fmt"
	"log"
	"os"

	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"go-micro.dev/v5/transport/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/tihe/susi-auth-service/internal/handler"
	"github.com/tihe/susi-auth-service/internal/repository"
	"github.com/tihe/susi-auth-service/internal/service"
	"github.com/tihe/susi-proto/auth"
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

	// Service configuration
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "auth-service"
	}

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

	// Graceful shutdown setup
	service := micro.NewService(
		micro.Name(serviceName),
		micro.Registry(consul.NewConsulRegistry(registry.Addrs(consulURL))),
		micro.Transport(grpc.NewTransport()),
		micro.AfterStop(func() error {
			// TODO: add graceful shutdown process
			log.Println("Auth service exiting")
			return nil
		}),
	)

	service.Init()

	auth.RegisterAuthServiceHandler(service.Server(), authHanlder)

	if err := service.Run(); err != nil {
		log.Printf("Error %s: %v", serviceName, err)
	}
}
