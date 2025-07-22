package models

import "time"

type Apartment struct {
	ID         int       `json:"id"`
	Address    string    `json:"address"`
	BaseRent   int       `json:"base_rent"`
	LandLordID int       `json:"land_lord_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
} 