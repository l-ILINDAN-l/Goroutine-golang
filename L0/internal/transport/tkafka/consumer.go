package tkafka

import (
	"L0/internal/app"
	"L0/internal/domain"
	"L0/internal/metrics"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
	"time"
)

var ErrInvalidMessageData = errors.New("invalid message data")

type Topics []string

type Consumer struct {
	kafkaConsumer *kafka.Consumer
	orderService  app.OrderService
	logger        *logrus.Entry
}

func NewConsumer(cfg *kafka.ConfigMap, service app.OrderService, logger *logrus.Entry) (*Consumer, error) {
	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		logger.Errorf("error creating tkafka consumer: %v", err)
		return nil, err
	}

	return &Consumer{
		kafkaConsumer: c,
		orderService:  service,
		logger:        logger,
	}, nil
}

func (c *Consumer) handleSingleMessage(ctx context.Context, msg *kafka.Message) error {
	order := domain.Order{}
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		metrics.KafkaMessagesInvalidTotal.Inc()
		return fmt.Errorf("unmarshal error (bad message): %w", err)
	}
	if order.Delivery == nil || order.Payment == nil {
		metrics.KafkaMessagesInvalidTotal.Inc()
		c.logger.Errorf("invalid order data: delivery or payment is null for order %s", order.OrderUID)
		return fmt.Errorf("%w: delivery or payment is null for order %s", ErrInvalidMessageData, order.OrderUID)
	}
	if err := c.orderService.ProcessNewOrder(ctx, &order); err != nil {
		return fmt.Errorf("service processing error: %w", err)
	}
	return nil
}

func (c *Consumer) StartConsuming(ctx context.Context, subscribingTopic Topics) {
	err := c.kafkaConsumer.SubscribeTopics(subscribingTopic, nil)
	if err != nil {
		c.logger.Errorf("error subscribing to Topics: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.kafkaConsumer.ReadMessage(1 * time.Second)
			if err == nil {
				metrics.KafkaMessagesReceivedTotal.Inc()
				if err = c.handleSingleMessage(ctx, msg); err != nil {
					if errors.Is(err, ErrInvalidMessageData) {
						c.logger.Warn("Skipping and committing bad message...")
						_, _ = c.kafkaConsumer.CommitMessage(msg)
						continue
					}
					c.logger.WithField("order_uid", msg.Key).Errorf("Failed to handle message: %v", err)
					continue
				} else {
					_, err = c.kafkaConsumer.CommitMessage(msg)
					if err != nil {
						c.logger.Errorf("Failed to commit message: %v", err)
					}
				}

			} else if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
				continue
			} else {
				c.logger.Errorf("error reading message: %v", err)
			}
		}
	}
}

func (c *Consumer) Close() {
	if err := c.kafkaConsumer.Close(); err != nil {
		c.logger.Errorf("error closing tkafka consumer: %v", err)
	}
}
