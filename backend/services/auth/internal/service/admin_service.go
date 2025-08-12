package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/tihe/susi-auth-service/internal/model"
	"github.com/tihe/susi-auth-service/internal/repository"
	"github.com/tihe/susi-shared/events"
)

type AdminService interface {
	CreateAdmin(ctx context.Context, admin *model.Admin) (*model.Admin, error)
	GetAdminByID(ctx context.Context, id int) (*model.Admin, error)
	UpdateAdmin(ctx context.Context, admin *model.Admin) (*model.Admin, error)
	DeleteAdmin(ctx context.Context, id int) error
	ListAdmins(ctx context.Context, page, limit int) ([]*model.Admin, int64, error)
}

type adminService struct {
	repo          repository.AdminRepository
	kafkaProducer *events.KafkaProducer
}

func NewAdminService(repo repository.AdminRepository, producer *events.KafkaProducer) AdminService {
	return &adminService{
		repo:          repo,
		kafkaProducer: producer,
	}
}

func (s *adminService) CreateAdmin(ctx context.Context, admin *model.Admin) (*model.Admin, error) {
	if err := s.validateAdminInput(*admin); err != nil {
		return nil, err
	}

	existingAdmin, _ := s.repo.GetByEmail(ctx, admin.Email)
	if existingAdmin != nil {
		return nil, errors.New("admin with email already exists")
	}

	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *adminService) GetAdminByID(ctx context.Context, id int) (*model.Admin, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adminService) UpdateAdmin(ctx context.Context, admin *model.Admin) (*model.Admin, error) {
	if err := s.validateAdminInput(*admin); err != nil {
		return nil, err
	}

	existingAdmin, err := s.repo.GetByID(ctx, admin.ID)
	if err != nil {
		return nil, err
	}

	if admin.Email != existingAdmin.Email {
		existingEmailAdmin, _ := s.repo.GetByEmail(ctx, admin.Email)
		if existingEmailAdmin != nil && existingEmailAdmin.ID != admin.ID {
			return nil, errors.New("admin with this email already exists")
		}
	}

	if err := s.repo.Update(ctx, admin); err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *adminService) DeleteAdmin(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *adminService) ListAdmins(ctx context.Context, page, limit int) ([]*model.Admin, int64, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	return s.repo.List(ctx, offset, limit)
}

func (s *adminService) validateAdminInput(admin model.Admin) error {
	if strings.TrimSpace(admin.Name) == "" {
		return errors.New("name is required")
	}

	if len(strings.TrimSpace(admin.Name)) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if strings.TrimSpace(admin.Email) == "" {
		return errors.New("email is required")
	}

	if !s.isValidEmail(admin.Email) {
		return errors.New("invalid email format")
	}

	return nil
}

func (s *adminService) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
