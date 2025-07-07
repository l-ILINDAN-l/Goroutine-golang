package main

import (
	"L0/internal/app"
	"L0/internal/cache"
	"L0/internal/config"
	"L0/internal/repository"
	"L0/internal/transport/thttp"
	"L0/internal/transport/tkafka"
	"context"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("service stopped with an error: %v", err)
	}
}

func run() error {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg := config.MustLoad()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	postgresRepo, err := repository.NewPostgresRepository(ctx, &cfg.Postgres, logger.WithField("component", "repository"))
	if err != nil {
		logger.Fatalf("failed to create postgres repository: %v", err)

		return err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatalf("failed to connect to redis: %v", err)
		return err
	}

	cacheRepo := cache.NewRedisCache(postgresRepo, redisClient, logger.WithField("component", "cache"))

	orderService := app.NewOrderService(cacheRepo, cacheRepo, logger.WithField("component", "service"))

	kafkaConsumer, err := tkafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(cfg.Kafka.Brokers, ","),
		"group.id":           "order_service_consumer",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}, orderService, logger.WithField("component", "kafka_consumer"))

	if err != nil {
		logger.Fatalf("failed to create tkafka consumer: %v", err)
		return err
	}

	httpServer := thttp.NewHTTPServer(gin.Default(), orderService, logger.WithField("component", "http_server"))
	httpServer.SetupRoutes()

	go func() {
		err := orderService.WarmUpCache(ctx, 1000)
		if err != nil {
			logger.Errorf("failed to warm up cache: %v", err)
		}
	}()

	go kafkaConsumer.StartConsuming(ctx, tkafka.Topics{cfg.Kafka.Topic})

	go func() {
		if err := httpServer.Run(ctx, cfg.Server.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("thttp server error: %v", err)
		}
	}()

	logger.Info("service started")

	<-ctx.Done()

	logger.Info("shutting down service...")

	kafkaConsumer.Close()
	cacheRepo.Close()
	postgresRepo.Close()

	logger.Info("service stopped")
	return nil
}
