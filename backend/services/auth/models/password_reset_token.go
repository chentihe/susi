package models

import "time"

type PasswordResetToken struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	AdminID   int       `json:"admin_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
