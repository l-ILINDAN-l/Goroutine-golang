package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

// LoggingMiddleware creates a gin middleware for logging HTTP requests
func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		entry := logger.WithFields(logrus.Fields{
			"statusCode": statusCode,
			"latency":    latency,
			"clientIP":   clientIP,
			"method":     method,
			"path":       path,
		})

		if statusCode >= 500 {
			entry.Error("request failed with a server error")
		} else if statusCode >= 400 {
			entry.Warn("request ended with a client error")
		} else {
			entry.Info("request has been successfully processed")
		}
	}
}
