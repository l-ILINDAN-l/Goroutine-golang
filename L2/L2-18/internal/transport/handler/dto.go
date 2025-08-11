package handler

import "github.com/google/uuid"

// CreateEventRequest defines the structure for a new event creation request
type CreateEventRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	Date      string    `json:"date" binding:"required"`
	EventText string    `json:"event_text" binding:"required"`
}

// UpdateEventRequest defines the structure for an event update request
type UpdateEventRequest struct {
	EventID   uuid.UUID `json:"event_id" binding:"required"`
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	Date      string    `json:"date" binding:"required"`
	EventText string    `json:"event_text" binding:"required"`
}

// DeleteEventRequest defines the structure for an event deletion request
type DeleteEventRequest struct {
	EventID uuid.UUID `json:"event_id" binding:"required"`
	UserID  uuid.UUID `json:"user_id" binding:"required"`
}

// SuccessResponse defines the standard structure for a successful API response
type SuccessResponse struct {
	Result map[string]any `json:"result"`
}

// ErrorResponse defines the standard structure for an error API response
type ErrorResponse struct {
	Error string `json:"error"`
}
