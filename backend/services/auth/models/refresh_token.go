package models

import "time"

type RefreshToken struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	AdminID   int       `json:"admin_id"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} 