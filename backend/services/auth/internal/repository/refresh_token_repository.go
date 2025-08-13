package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tihe/susi-auth-service/internal/model"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *model.RefreshToken) error
	Update(ctx context.Context, refreshToken *model.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	Delete(ctx context.Context, token string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, refreshToken *model.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(refreshToken).Error; err != nil {
		return err
	}

	return nil
}

func (r *refreshTokenRepository) Update(ctx context.Context, refreshToken *model.RefreshToken) error {
	result := r.db.WithContext(ctx).Save(refreshToken)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}
	return nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	if err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) Delete(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Delete(&model.RefreshToken{}, token)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}
	return nil
}
