package cache

import (
	"calendar/internal/domain/event"
	"context"
	"github.com/google/uuid"
)

// Cache defines the contract for a cache layer for events
type Cache interface {
	// Set stores an event in the cache
	Set(ctx context.Context, event *event.Event) error
	// Get retrieves an event from the cache by its ID
	Get(ctx context.Context, eventID uuid.UUID) (*event.Event, bool, error)
	// Delete removes an event from the cache by its ID
	Delete(ctx context.Context, eventID uuid.UUID) error
}
