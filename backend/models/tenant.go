package models

import "time"

type Tenant struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	RoomID     int       `json:"room_id"`
	LeaseStart time.Time `json:"lease_start"`
	LeaseEnd   time.Time `json:"lease_end"`
	MonthlyRent int      `json:"monthly_rent"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
} 