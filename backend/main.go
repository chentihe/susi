package main

import (
	"log"
	"os"
	"strings"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/tihe/susi/backend/app"
	"github.com/tihe/susi/backend/events"
	"github.com/tihe/susi/backend/models"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	apartmentRepo := models.NewApartmentImpl(db)

	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	producer := events.NewKafkaProducer(events.KafkaConfig{
		Brokers: kafkaBrokers,
		Topic:   kafkaTopic,
	})
	defer producer.Close()

	ctx := app.NewAppContext(db, producer, apartmentRepo)
	router := app.NewAppRouter(ctx)

	router.Engine.Run(":8080")
} 