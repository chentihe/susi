package models

import "time"

type Admin struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"unique;not null"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
} 