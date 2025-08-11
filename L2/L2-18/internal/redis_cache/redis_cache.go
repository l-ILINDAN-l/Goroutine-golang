package rediscache

import (
	"calendar/internal/config"
	"calendar/internal/domain/event"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

// RedisCache is a Redis implementation of the cache.Cache interface
type RedisCache struct {
	client *redis.Client
	logger *logrus.Logger
	ttl    time.Duration
}

// New creates a new instance of RedisCache
func New(cfg *config.RedisConfig, logger *logrus.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		logger: logger,
		ttl:    5 * time.Minute,
	}, nil
}

// Close closes the connection to Redis
func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) generateKey(eventID uuid.UUID) string {
	return fmt.Sprintf("event:%s", eventID.String())
}

// Set stores an event in the cache
func (c *RedisCache) Set(ctx context.Context, event *event.Event) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshal event: %w", err)
	}

	key := c.generateKey(event.ID)

	return c.client.Set(ctx, key, string(eventBytes), c.ttl).Err()
}

// Get retrieves an event from the cache by its ID
func (c *RedisCache) Get(ctx context.Context, eventID uuid.UUID) (*event.Event, bool, error) {
	key := c.generateKey(eventID)

	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var ev event.Event
	if err := json.Unmarshal(value, &ev); err != nil {
		return nil, false, fmt.Errorf("error unmarshal event: %w", err)
	}
	return &ev, true, nil
}

// Delete removes an event from the cache
func (c *RedisCache) Delete(ctx context.Context, eventID uuid.UUID) error {
	key := c.generateKey(eventID)
	return c.client.Del(ctx, key).Err()
}
