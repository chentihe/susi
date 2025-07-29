package services

import (
	"time"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/events"
)

type TenantService interface {
	CreateTenant(tenant *models.Tenant) error
	// Add more business methods as needed
}

type TenantServiceImpl struct {
	Repo          models.TenantRepository
	KafkaProducer *events.KafkaProducer
}

func NewTenantService(repo models.TenantRepository, producer *events.KafkaProducer) TenantService {
	return &TenantServiceImpl{
		Repo: repo,
		KafkaProducer: producer,
	}
}

func (s *TenantServiceImpl) CreateTenant(tenant *models.Tenant) error {
	err := s.Repo.Create(tenant)
	if err != nil {
		return err
	}
	event := events.Event{
		Type:      "TenantCreated",
		Payload:   tenant,
		Timestamp: time.Now(),
	}
	s.KafkaProducer.Publish(event)
	return nil
} 