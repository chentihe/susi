package services

import (
	"time"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/events"
)

type RoomService interface {
	CreateRoom(room *models.Room) error
	// Add more business methods as needed
}

type RoomServiceImpl struct {
	Repo          models.RoomRepository
	KafkaProducer *events.KafkaProducer
}

func NewRoomService(repo models.RoomRepository, producer *events.KafkaProducer) RoomService {
	return &RoomServiceImpl{
		Repo: repo,
		KafkaProducer: producer,
	}
}

func (s *RoomServiceImpl) CreateRoom(room *models.Room) error {
	err := s.Repo.Create(room)
	if err != nil {
		return err
	}
	event := events.Event{
		Type:      events.EventRoomRented, // Or a more appropriate event type
		Payload:   room,
		Timestamp: time.Now(),
	}
	s.KafkaProducer.Publish(event)
	return nil
} 