package thttp

import (
	"L0/internal/repository"
	"errors"
	"github.com/gin-gonic/gin"

	"net/http"
)

func (s *Server) getOrderHandler(c *gin.Context) {

	orderUID := c.Param("order_uid")
	order, err := s.orderService.GetOrderByUID(c.Request.Context(), orderUID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) || errors.Is(err, repository.ErrShardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		} else {
			s.logger.Errorf("failed to get order by uid %s: %v", orderUID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	orderResponse := toOrderResponse(order)
	c.JSON(http.StatusOK, orderResponse)
}
