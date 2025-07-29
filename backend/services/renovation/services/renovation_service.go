package services

import (
	"time"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/events"
)

type RenovationService interface {
	CreateRenovation(renovation *models.Renovation) error
	// Add more business methods as needed
}

type RenovationServiceImpl struct {
	Repo          models.RenovationRepository
	KafkaProducer *events.KafkaProducer
}

func NewRenovationService(repo models.RenovationRepository, producer *events.KafkaProducer) RenovationService {
	return &RenovationServiceImpl{
		Repo: repo,
		KafkaProducer: producer,
	}
}

func (s *RenovationServiceImpl) CreateRenovation(renovation *models.Renovation) error {
	err := s.Repo.Create(renovation)
	if err != nil {
		return err
	}
	event := events.Event{
		Type:      "RenovationCreated",
		Payload:   renovation,
		Timestamp: time.Now(),
	}
	s.KafkaProducer.Publish(event)
	return nil
} 