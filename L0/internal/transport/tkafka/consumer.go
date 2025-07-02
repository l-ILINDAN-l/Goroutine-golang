package tkafka

import (
	"L0/internal/app"
	"L0/internal/domain"
	"L0/internal/metrics"
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
	"time"
)

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

				orderBytes := msg.Value
				order := domain.Order{}
				if err = json.Unmarshal(orderBytes, &order); err != nil {
					c.logger.Errorf("failed to unmarshal order: %v", err)
					metrics.KafkaMessagesInvalidTotal.Inc()
					continue
				}
				if order.Delivery == nil || order.Payment == nil {
					c.logger.Errorf("invalid order data: delivery or payment is null for order %s", order.OrderUID)
					if _, err = c.kafkaConsumer.CommitMessage(msg); err != nil {
						c.logger.Errorf("failed to commit bad message: %v", err)
					}
					continue
				}
				if err = c.orderService.ProcessNewOrder(ctx, &order); err != nil {
					c.logger.Errorf("failed to save order: %v", err)
					continue
				}
				if _, err = c.kafkaConsumer.CommitMessage(msg); err != nil {
					c.logger.Errorf("failed to commit message: %v", err)
					metrics.KafkaMessagesInvalidTotal.Inc()
					continue
				}
				c.logger.Infof("consumed order: %v", order)

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
