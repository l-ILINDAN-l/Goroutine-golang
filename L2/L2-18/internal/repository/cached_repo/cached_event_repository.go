package cachedrepo

import (
	"calendar/internal/domain/cache"
	"calendar/internal/domain/event"
	"calendar/internal/domain/event_repository"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

// CachedEventRepository is a decorator for an EventRepository that adds a caching layer
type CachedEventRepository struct {
	mainRepo eventrepository.EventRepository
	cache    cache.Cache
	logger   *logrus.Logger
}

// New creates a new instance of the cached repository
func New(mainRepo eventrepository.EventRepository, cache cache.Cache, logger *logrus.Logger) *CachedEventRepository {
	return &CachedEventRepository{
		mainRepo: mainRepo,
		cache:    cache,
		logger:   logger,
	}
}

// Create persists the event and then caches it
func (r *CachedEventRepository) Create(ctx context.Context, event *event.Event) error {
	if err := r.mainRepo.Create(ctx, event); err != nil {
		return err
	}
	return r.cache.Set(ctx, event)
}

// Update modifies the event in the main repository and then invalidates/updates the cache
func (r *CachedEventRepository) Update(ctx context.Context, event *event.Event) error {
	if err := r.mainRepo.Update(ctx, event); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, event.ID); err != nil {
		return err
	}

	return r.cache.Set(ctx, event)
}

// Delete removes the event from the main repository and the cache
func (r *CachedEventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	if err := r.mainRepo.Delete(ctx, eventID); err != nil {
		return err
	}
	return r.cache.Delete(ctx, eventID)
}

// FindByID first checks the cache for an event, falling back to the main repository on a miss
func (r *CachedEventRepository) FindByID(ctx context.Context, eventID uuid.UUID) (*event.Event, error) {
	cachedEvent, found, err := r.cache.Get(ctx, eventID)
	if err != nil {
		r.logger.Errorf("error fetch event by id %v", eventID)
	}
	if found {
		r.logger.Infof("CACHE HIT event by id %v", eventID)
		return cachedEvent, nil
	}

	r.logger.Infof("CACHE HIT event by id %v", eventID)

	dbEvent, err := r.mainRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(ctx, dbEvent); err != nil {
		r.logger.Errorf("error cache set event by id %v after FindByID", eventID)
	}

	return dbEvent, nil
}

// FindByDateRange proxies the request directly to the main repository as date range queries are not cached
func (r *CachedEventRepository) FindByDateRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*event.Event, error) {
	return r.mainRepo.FindByDateRange(ctx, userID, startTime, endTime)
}

// GetLatest retrieves the latest events and uses them to warm up the cache
func (r *CachedEventRepository) GetLatest(ctx context.Context, limit uint) ([]*event.Event, error) {
	r.logger.Infof("WARMING up the CACHE wiht LIMIT %d", limit)
	events, err := r.mainRepo.GetLatest(ctx, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	for _, ev := range events {
		if err := r.cache.Set(ctx, ev); err != nil {
			r.logger.Errorf("error CACHE SET event %v for WARMING up cahce", ev)
		}
	}
	return events, nil
}
