package cache

import (
	"L0/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

type RedisCache struct {
	redisClient *redis.Client
	nextLayer   domain.OrderRepository
	logger      *logrus.Entry
}

func NewRedisCache(repo domain.OrderRepository, redisClient *redis.Client, logger *logrus.Entry) *RedisCache {
	return &RedisCache{
		redisClient: redisClient,
		nextLayer:   repo,
		logger:      logger,
	}
}

func createKey(uid uuid.UUID) string {
	return fmt.Sprintf("order:%v", uid)
}

func (c *RedisCache) getOrderFromDatabaseAndCache(ctx context.Context, uid uuid.UUID, log *logrus.Entry) (*domain.Order, error) {
	order, err := c.nextLayer.GetByUID(ctx, uid)
	if err != nil {
		log.Errorf("failed to get order: %v", err)
		return nil, err
	}

	jsonOrderBytes, err := json.Marshal(order)
	if err != nil {
		log.Errorf("failed to marshal order for cashing, returning order without caching:: %v", err)
		return order, nil
	}
	orderKey := createKey(uid)
	if err = c.redisClient.Set(ctx, orderKey, jsonOrderBytes, 15*time.Minute).Err(); err != nil {
		log.Errorf("failed to set order into cache: %v", err)
	}

	return order, nil
}

func (c *RedisCache) GetByUID(ctx context.Context, uid uuid.UUID) (*domain.Order, error) {
	log := c.logger.WithFields(logrus.Fields{
		"uid": uid,
	})
	orderKey := createKey(uid)
	jsonData, err := c.redisClient.Get(ctx, orderKey).Bytes()

	if err == nil {
		log.Info("cache hit")

		var order domain.Order

		if err = json.Unmarshal(jsonData, &order); err != nil {
			log.Errorf("failed to unmarshal order: %v", err)
			return c.getOrderFromDatabaseAndCache(ctx, uid, log)
		}
		return &order, nil
	} else if !errors.Is(err, redis.Nil) {
		log.Errorf("failed to get order: %v, falling back to database", err)
	} else {
		log.Info("cache miss, falling back to database")
	}

	return c.getOrderFromDatabaseAndCache(ctx, uid, log)
}

func (c *RedisCache) Save(ctx context.Context, order *domain.Order) error {
	log := c.logger.WithFields(logrus.Fields{
		"uid": order.OrderUID,
	})
	if err := c.nextLayer.Save(ctx, order); err != nil {
		log.Errorf("failed to save order: %v", err)
		return err
	} else {
		jsonOrderBytes, err := json.Marshal(order)
		if err != nil {
			log.Errorf("failed to marshal order for saving: %v", err)
			return nil
		}
		if err = c.redisClient.Set(ctx, createKey(order.OrderUID), jsonOrderBytes, 15*time.Minute).Err(); err != nil {
			log.Errorf("failed to save order into cache: %v", err)
		} else {
			log.Info("cache updated successfully after save")
		}
		return nil
	}
}

func (c *RedisCache) GetLatest(ctx context.Context, limit uint) ([]*domain.Order, error) {
	return c.nextLayer.GetLatest(ctx, limit)
}

func (c *RedisCache) Set(ctx context.Context, order *domain.Order) error {
	log := c.logger.WithField("order_uid", order.OrderUID.String())

	jsonOrderBytes, err := json.Marshal(order)
	if err != nil {
		log.Errorf("failed to marshal order for caching: %v", err)
		return nil
	}

	orderKey := createKey(order.OrderUID)
	if err := c.redisClient.Set(ctx, orderKey, jsonOrderBytes, 15*time.Minute).Err(); err != nil {
		log.Errorf("failed to set order into cache: %v", err)
	} else {
		log.Info("cache entry set successfully")
	}

	return nil
}

func (c *RedisCache) Get(ctx context.Context, uid uuid.UUID) (*domain.Order, bool, error) {
	log := c.logger.WithField("order_uid", uid.String())
	orderKey := createKey(uid)

	jsonData, err := c.redisClient.Get(ctx, orderKey).Bytes()
	if err == nil {

		log.Info("cache hit")
		var order domain.Order
		if err := json.Unmarshal(jsonData, &order); err != nil {
			log.Errorf("failed to unmarshal corrupted cache data: %v", err)
			return nil, false, nil
		}
		return &order, true, nil
	}

	if errors.Is(err, redis.Nil) {
		log.Info("cache miss")
		return nil, false, nil
	}

	// Настоящая ошибка Redis
	log.Errorf("failed to get from cache: %v", err)
	return nil, false, err
}
