package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tihe/susi-auth-service/internal/model"
)

type AdminRepository interface {
	Create(ctx context.Context, admin *model.Admin) error
	GetByID(ctx context.Context, id int) (*model.Admin, error)
	GetByEmail(ctx context.Context, email string) (*model.Admin, error)
	Update(ctx context.Context, admin *model.Admin) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*model.Admin, int64, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (a *adminRepository) Create(ctx context.Context, admin *model.Admin) error {
	if err := a.db.WithContext(ctx).Create(admin).Error; err != nil {
		return err
	}
	return nil
}

func (a *adminRepository) GetByID(ctx context.Context, id int) (*model.Admin, error) {
	var admin model.Admin
	if err := a.db.WithContext(ctx).First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (a *adminRepository) GetByEmail(ctx context.Context, email string) (*model.Admin, error) {
	var admin model.Admin
	if err := a.db.WithContext(ctx).First(&admin, email).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (a *adminRepository) Update(ctx context.Context, admin *model.Admin) error {
	result := a.db.WithContext(ctx).Save(admin)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("admin not found")
	}
	return nil
}

func (a *adminRepository) Delete(ctx context.Context, id int) error {
	result := a.db.WithContext(ctx).Delete(&model.Admin{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("admin not found")
	}
	return nil
}

func (a *adminRepository) List(ctx context.Context, offset, limit int) ([]*model.Admin, int64, error) {
	var admins []*model.Admin
	var total int64
	if err := a.db.WithContext(ctx).Model(&model.Admin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := a.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&admins).Error; err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}
