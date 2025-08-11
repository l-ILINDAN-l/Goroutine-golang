package main

import (
	"calendar/internal/app"
	"calendar/internal/config"
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Warnf("couldn't upload .env file: %v", err)
	}

	// 2. Загружаем конфиг
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("configuration loading error: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatalf("application initialization error: %v", err)
	}

	go func() {
		if err := application.Run(); err != nil {
			logger.Fatalf("HTTP server startup error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Stop(ctx); err != nil {
		logger.Errorf("error stopping the app correctly: %v", err)
	}

	logger.Info("service has been successfully stopped")
}
