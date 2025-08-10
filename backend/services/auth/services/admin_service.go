package services

import (
	"time"

	"github.com/tihe/susi-auth-service/models"
	"github.com/tihe/susi-shared/events"
)

type AdminService interface {
	CreateAdmin(admin *models.Admin) error
	// Add more business methods as needed
}

type AdminServiceImpl struct {
	Repo          models.AdminRepository
	KafkaProducer *events.KafkaProducer
}

func NewAdminService(repo models.AdminRepository, producer *events.KafkaProducer) AdminService {
	return &AdminServiceImpl{
		Repo:          repo,
		KafkaProducer: producer,
	}
}

func (s *AdminServiceImpl) CreateAdmin(admin *models.Admin) error {
	err := s.Repo.Create(admin)
	if err != nil {
		return err
	}
	event := events.Event{
		Type:      "AdminRegistered",
		Payload:   admin,
		Timestamp: time.Now(),
	}
	s.KafkaProducer.Publish(event)
	return nil
}
