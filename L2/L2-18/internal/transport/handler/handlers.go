package handler

import (
	"calendar/internal/service"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// Handler holds dependencies for the HTTP handlers
type Handler struct {
	eventService service.EventService
}

// New creates a new Handler instance
func New(eventService service.EventService) *Handler {
	return &Handler{
		eventService: eventService,
	}
}

// RegisterRoutes sets up all the API routes for the application
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.POST("/create_event", h.createEvent)
		api.POST("/update_event", h.updateEvent)
		api.POST("/delete_event", h.deleteEvent)
		api.GET("/events_for_day", h.eventsForDay)
		api.GET("/events_for_week", h.eventsForWeek)
		api.GET("/events_for_month", h.eventsForMonth)
	}
}

func (h *Handler) createEvent(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid data format, expect YYYY-MM-DD"})
		return
	}

	event, err := h.eventService.CreateEvent(c.Request.Context(), req.UserID, date, req.EventText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error server"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Result: gin.H{"event": event}})
}

func (h *Handler) updateEvent(c *gin.Context) {
	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date format, expect YYYY-MM-DD"})
		return
	}

	event, err := h.eventService.UpdateEvent(c.Request.Context(), req.EventID, req.UserID, date, req.EventText)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error server"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Result: gin.H{"event": event}})
}

func (h *Handler) deleteEvent(c *gin.Context) {
	var req DeleteEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.eventService.DeleteEvent(c.Request.Context(), req.EventID, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error server"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Result: gin.H{"status": "OK"}})
}

func (h *Handler) eventsForDay(c *gin.Context) {
	userIDStr := c.Query("user_id")
	dateStr := c.Query("date")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid format user_id"})
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date format, expect YYYY-MM-DD"})
		return
	}

	events, err := h.eventService.GetEventsForDay(c.Request.Context(), userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error server"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Result: gin.H{"events": events}})
}

func (h *Handler) eventsForWeek(c *gin.Context) {
	userIDStr := c.Query("user_id")
	dateStr := c.Query("date")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid format user_id"})
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date format, expect YYYY-MM-DD"})
		return
	}

	events, err := h.eventService.GetEventsForWeek(c.Request.Context(), userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error server"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Result: gin.H{"events": events}})
}

func (h *Handler) eventsForMonth(c *gin.Context) {
	userIDStr := c.Query("user_id")
	dateStr := c.Query("date")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid format user_id"})
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date format, expect YYYY-MM-DD"})
		return
	}

	events, err := h.eventService.GetEventsForMonth(c.Request.Context(), userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error server"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Result: gin.H{"events": events}})
}
