package models

import "time"

type Room struct {
	ID          int       `json:"id"`
	ApartmentID int       `json:"apartment_id"`
	Name        string    `json:"name"`
	MonthlyRent int       `json:"monthly_rent"`
	Status      string    `json:"status"` // e.g., vacant, occupied
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} 