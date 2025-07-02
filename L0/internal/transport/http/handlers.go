package http

import (
	"L0/internal/repository"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func (s *Server) getOrderHandler(c *gin.Context) {

	orderUIDstr := c.Param("order_uid")
	orderUID, err := uuid.Parse(orderUIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order_uid format:"})
		return
	}
	order, err := s.orderService.GetOrderByUID(c.Request.Context(), orderUID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		} else {
			s.logger.Errorf("failed to get order by uid %s: %v", orderUIDstr, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	orderResponse := toOrderResponse(order)
	c.JSON(http.StatusOK, orderResponse)
}
