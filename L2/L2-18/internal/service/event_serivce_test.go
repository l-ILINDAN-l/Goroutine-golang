package service

import (
	"calendar/internal/domain/event"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
	"time"
)

// MockEventRepository is a mock type for the EventRepository interface.
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, e *event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

// Update implements the EventRepository interface.
func (m *MockEventRepository) Update(ctx context.Context, e *event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

// Delete implements the EventRepository interface.
func (m *MockEventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

// FindByID implements the EventRepository interface.
func (m *MockEventRepository) FindByID(ctx context.Context, eventID uuid.UUID) (*event.Event, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*event.Event), args.Error(1)
}

// FindByDateRange implements the EventRepository interface.
func (m *MockEventRepository) FindByDateRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*event.Event, error) {
	args := m.Called(ctx, userID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*event.Event), args.Error(1)
}

// GetLatest implements the EventRepository interface.
func (m *MockEventRepository) GetLatest(ctx context.Context, limit uint) ([]*event.Event, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*event.Event), args.Error(1)
}

func TestEventService_CreateEvent(t *testing.T) {
	t.Run("should create event successfully", func(t *testing.T) {
		mockRepo := new(MockEventRepository)
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		eventService := New(mockRepo, logger)

		userID := uuid.New()
		eventDate := time.Now()
		eventText := "Test Event"

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil).Once()

		// Act: Call the method we are testing
		createdEvent, err := eventService.CreateEvent(context.Background(), userID, eventDate, eventText)

		// Assert: Check the results
		assert.NoError(t, err)
		assert.NotNil(t, createdEvent)
		assert.Equal(t, userID, createdEvent.UserID)
		assert.Equal(t, eventText, createdEvent.EventText)
		assert.NotEqual(t, uuid.Nil, createdEvent.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockEventRepository)
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		eventService := New(mockRepo, logger)

		repoError := errors.New("database error")

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*event.Event")).Return(repoError).Once()

		createdEvent, err := eventService.CreateEvent(context.Background(), uuid.New(), time.Now(), "Test Event")

		assert.Error(t, err)
		assert.Nil(t, createdEvent)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestEventService_UpdateEvent(t *testing.T) {
	t.Run("should update event successfully", func(t *testing.T) {
		mockRepo := new(MockEventRepository)
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		eventService := New(mockRepo, logger)

		eventID := uuid.New()
		userID := uuid.New()

		existingEvent := &event.Event{
			ID:        eventID,
			UserID:    userID,
			Date:      time.Now(),
			EventText: "Old Text",
		}

		mockRepo.On("FindByID", mock.Anything, eventID).Return(existingEvent, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil).Once()

		updatedEvent, err := eventService.UpdateEvent(context.Background(), eventID, userID, time.Now().Add(time.Hour), "New Text")

		assert.NoError(t, err)
		assert.NotNil(t, updatedEvent)
		assert.Equal(t, "New Text", updatedEvent.EventText)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if event not found", func(t *testing.T) {
		mockRepo := new(MockEventRepository)
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		eventService := New(mockRepo, logger)

		eventID := uuid.New()

		mockRepo.On("FindByID", mock.Anything, eventID).Return(nil, sql.ErrNoRows).Once()

		updatedEvent, err := eventService.UpdateEvent(context.Background(), eventID, uuid.New(), time.Now(), "New Text")

		assert.Error(t, err)
		assert.Nil(t, updatedEvent)
		assert.ErrorIs(t, err, sql.ErrNoRows)

		mockRepo.AssertExpectations(t)
	})
}
