package http

import (
	"L0/internal/app"
	"L0/internal/metrics"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	serverEngine *gin.Engine
	orderService app.OrderService
	logger       *logrus.Entry
}

func NewHTTPServer(serverEngine *gin.Engine, orderService app.OrderService, logger *logrus.Entry) *Server {
	return &Server{
		serverEngine: serverEngine,
		orderService: orderService,
		logger:       logger,
	}
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		path := c.FullPath()
		method := c.Request.Method
		statusCode := strconv.Itoa(c.Writer.Status())

		metrics.HttpRequestDuration.Observe(duration.Seconds())
		metrics.HttpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()

	}
}

func (s *Server) SetupRoutes() {
	s.serverEngine.Use(PrometheusMiddleware())
	s.serverEngine.GET("/order/:order_uid", s.getOrderHandler)
	s.serverEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.serverEngine.StaticFS("/", http.Dir("./static"))
}

func (s *Server) Run(ctx context.Context, port string) error {
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: s.serverEngine,
	}

	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Errorf("HTTP server shutdown error: %v", err)
		}
	}()

	s.logger.Infof("HTTP server started on port %s", port)
	return httpServer.ListenAndServe()
}
