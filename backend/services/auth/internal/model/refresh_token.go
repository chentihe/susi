package model

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"unique;not null;type:varchar(255);uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Revoked   bool      `json:"revoked" gorm:"deault:false"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && time.Now().Before(rt.ExpiresAt)
}
