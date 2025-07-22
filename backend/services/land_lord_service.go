package services

import (
	"time"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/events"
)

type LandLordService interface {
	CreateLandLord(landLord *models.LandLord) error
	// Add more business methods as needed
}

type LandLordServiceImpl struct {
	Repo          models.LandLordRepository
	KafkaProducer *events.KafkaProducer
}

func NewLandLordService(repo models.LandLordRepository, producer *events.KafkaProducer) LandLordService {
	return &LandLordServiceImpl{
		Repo: repo,
		KafkaProducer: producer,
	}
}

func (s *LandLordServiceImpl) CreateLandLord(landLord *models.LandLord) error {
	err := s.Repo.Create(landLord)
	if err != nil {
		return err
	}
	event := events.Event{
		Type:      "LandLordCreated",
		Payload:   landLord,
		Timestamp: time.Now(),
	}
	s.KafkaProducer.Publish(event)
	return nil
} 