package service

import (
	"calendar/internal/domain/event"
	"calendar/internal/domain/event_repository"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

// EventService encapsulates the business logic for managing events.
type EventService struct {
	repo   eventrepository.EventRepository
	logger *logrus.Logger
}

// New creates a new instance of the EventService.
func New(repo eventrepository.EventRepository, logger *logrus.Logger) *EventService {
	return &EventService{
		repo:   repo,
		logger: logger,
	}
}

// CreateEvent handles the business logic for creating a new event
func (s *EventService) CreateEvent(ctx context.Context, userID uuid.UUID, date time.Time, text string) (*event.Event, error) {
	newEvent := event.New(userID, date, text)

	if err := s.repo.Create(ctx, newEvent); err != nil {
		return nil, err
	}

	s.logger.Infof("event created ID: %s", newEvent.ID)
	return newEvent, nil
}

// UpdateEvent handles the business logic for updating an existing event
func (s *EventService) UpdateEvent(ctx context.Context, eventID, userID uuid.UUID, date time.Time, text string) (*event.Event, error) {
	existingEvent, err := s.repo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if existingEvent.UserID != userID {
		return nil, fmt.Errorf("user doesn`t have the right to edit")
	}

	existingEvent.Date = date
	existingEvent.EventText = text

	if err := s.repo.Update(ctx, existingEvent); err != nil {
		return nil, err
	}

	s.logger.Infof("event updated ID: %s", eventID)
	return existingEvent, nil
}

// DeleteEvent handles the business logic for deleting an event
func (s *EventService) DeleteEvent(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) error {
	existingEvent, err := s.repo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}

	if existingEvent.UserID != userID {
		return fmt.Errorf("user doesn`t have the right to delete")
	}

	if err := s.repo.Delete(ctx, eventID); err != nil {
		return err
	}
	s.logger.Infof("event deleted ID: %s", eventID)
	return nil
}

// GetEventsForDay retrieves all events for a specific user on a given day
func (s *EventService) GetEventsForDay(ctx context.Context, userID uuid.UUID, date time.Time) ([]*event.Event, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
	return s.repo.FindByDateRange(ctx, userID, startOfDay, endOfDay)
}

// GetEventsForWeek retrieves all events for a specific user for the week of a given day
func (s *EventService) GetEventsForWeek(ctx context.Context, userID uuid.UUID, date time.Time) ([]*event.Event, error) {
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := date.AddDate(0, 0, 1-weekday)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, date.Location())

	endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-1 * time.Nanosecond)

	return s.repo.FindByDateRange(ctx, userID, startOfWeek, endOfWeek)
}

// GetEventsForMonth retrieves all events for a specific user for the month of a given day
func (s *EventService) GetEventsForMonth(ctx context.Context, userID uuid.UUID, date time.Time) ([]*event.Event, error) {
	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())

	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)

	return s.repo.FindByDateRange(ctx, userID, startOfMonth, endOfMonth)
}
