package app

import (
	"database/sql"
	"github.com/tihe/susi/backend/events"
	"github.com/tihe/susi/backend/models"
)

type AppContext struct {
	DB             interface{} // keep for now, but not used directly
	KafkaProducer  *events.KafkaProducer
	ApartmentRepo  models.ApartmentRepository
}

func NewAppContext(db interface{}, producer *events.KafkaProducer, apartmentRepo models.ApartmentRepository) *AppContext {
	return &AppContext{
		DB:            db,
		KafkaProducer: producer,
		ApartmentRepo: apartmentRepo,
	}
} 