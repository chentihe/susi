package models

import "time"

type Renovation struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	ApartmentID int       `json:"apartment_id"`
	Description string    `json:"description"`
	TotalCost   int       `json:"total_cost"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} 