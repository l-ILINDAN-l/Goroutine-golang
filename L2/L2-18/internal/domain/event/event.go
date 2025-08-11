package event

import (
	"github.com/google/uuid"
	"time"
)

// Event represents the core domain model for a calendar event
type Event struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Date      time.Time `json:"date"`
	EventText string    `json:"event_text"`
}

// New creates a new Event instance with a generated unique ID
func New(userID uuid.UUID, date time.Time, eventText string) *Event {
	return &Event{
		ID:        uuid.New(),
		UserID:    userID,
		Date:      date,
		EventText: eventText,
	}
}
