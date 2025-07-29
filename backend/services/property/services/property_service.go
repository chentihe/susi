package services

import (
	"time"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/events"
)

type ApartmentService interface {
	CreateApartment(apartment *models.Apartment) error
	// Add more business methods as needed
}

type ApartmentServiceImpl struct {
	Repo          models.ApartmentRepository
	KafkaProducer *events.KafkaProducer
}

func NewApartmentService(repo models.ApartmentRepository, producer *events.KafkaProducer) ApartmentService {
	return &ApartmentServiceImpl{
		Repo: repo,
		KafkaProducer: producer,
	}
}

func (s *ApartmentServiceImpl) CreateApartment(apartment *models.Apartment) error {
	err := s.Repo.Create(apartment)
	if err != nil {
		return err
	}
	event := events.Event{
		Type:      events.EventApartmentCreated,
		Payload:   apartment,
		Timestamp: time.Now(),
	}
	s.KafkaProducer.Publish(event)
	return nil
} 