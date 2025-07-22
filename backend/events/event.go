package events

import "time"

// EventType constants
const (
	EventApartmentCreated = "ApartmentCreated"
	EventRoomRented      = "RoomRented"
	// Add more event types as needed
)

type Event struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
} 