package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/tihe/susi/backend/app"
	"github.com/tihe/susi/backend/events"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	producer := events.NewKafkaProducer(events.KafkaConfig{
		Brokers: kafkaBrokers,
		Topic:   kafkaTopic,
	})
	defer producer.Close()

	// Repositories
	apartmentRepo := models.NewApartmentImpl(db)
	roomRepo := models.NewRoomImpl(db)
	tenantRepo := models.NewTenantImpl(db)
	renovationRepo := models.NewRenovationImpl(db)
	landLordRepo := models.NewLandLordImpl(db)
	adminRepo := models.NewAdminImpl(db)

	// Services
	apartmentService := services.NewApartmentService(apartmentRepo, producer)
	roomService := services.NewRoomService(roomRepo, producer)
	tenantService := services.NewTenantService(tenantRepo, producer)
	renovationService := services.NewRenovationService(renovationRepo, producer)
	landLordService := services.NewLandLordService(landLordRepo, producer)
	adminService := services.NewAdminService(adminRepo, producer)

	ctx := app.NewAppContext(
		db,
		producer,
		apartmentService,
		roomService,
		tenantService,
		renovationService,
		landLordService,
		adminService,
	)

	router := app.NewAppRouter(ctx)

	// Graceful shutdown setup
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router.Engine,
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
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
} 