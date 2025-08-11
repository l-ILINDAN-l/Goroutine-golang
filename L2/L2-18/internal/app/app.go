package app

import (
	"calendar/internal/config"
	"calendar/internal/redis_cache"
	"calendar/internal/repository/cached_repo"
	"calendar/internal/repository/postgres_repo"
	"calendar/internal/service"
	"calendar/internal/transport/handler"
	"calendar/internal/transport/middleware"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// App is the main application structure, holding all dependencies
type App struct {
	cfg    *config.Config
	logger *logrus.Logger
	server *http.Server
}

// New creates and initializes a new App instance
func New(cfg *config.Config, logger *logrus.Logger) (*App, error) {

	dbRepo, err := postgresrepo.New(context.Background(), &cfg.Postgres, logger)
	if err != nil {
		return nil, err
	}

	redisCache, err := rediscache.New(&cfg.Redis, logger)
	if err != nil {
		return nil, err
	}

	cachedRepo := cachedrepo.New(dbRepo, redisCache, logger)
	eventService := service.New(cachedRepo, logger)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(gin.Recovery())

	apiHandler := handler.New(*eventService)
	apiHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
	}, nil
}

// Run starts the HTTP server. It is a blocking operation
func (a *App) Run() error {
	a.logger.Infof("HTTP-сервер запущен на порту %s", a.cfg.Server.Port)
	err := a.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop gracefully shuts down the HTTP server
func (a *App) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
