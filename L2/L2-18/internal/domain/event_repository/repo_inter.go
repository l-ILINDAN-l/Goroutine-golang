package eventrepository

import (
	"calendar/internal/domain/event"
	"context"
	"github.com/google/uuid"
	"time"
)

// EventRepository defines the contract for storing and retrieving event data
type EventRepository interface {
	// Create persists a new event
	Create(ctx context.Context, event *event.Event) error
	// Update modifies an existing event
	Update(ctx context.Context, event *event.Event) error
	// Delete removes an event by its ID
	Delete(ctx context.Context, eventID uuid.UUID) error
	// FindByID retrieves a single event by its unique ID
	FindByID(ctx context.Context, uid uuid.UUID) (*event.Event, error)
	// FindByDateRange retrieves all events for a specific user within a given time range
	FindByDateRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*event.Event, error)
	// GetLatest retrieves the most recent events up to a given limit, used for cache warming
	GetLatest(ctx context.Context, limit uint) ([]*event.Event, error)
}
