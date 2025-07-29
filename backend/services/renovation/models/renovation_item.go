package models

import "time"

type RenovationItem struct {
	ID             int                `json:"id" gorm:"primaryKey;autoIncrement"`
	RenovationID   int                `json:"renovation_id"`
	TypeID         int                `json:"type_id"`
	Type           RenovationItemType `json:"type" gorm:"foreignKey:TypeID;references:ID"`
	Cost           int                `json:"cost"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
} 