package app

import (
	"L0/internal/domain"
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type OrderService struct {
	repo  domain.OrderRepository
	cache domain.Cache

	logger *logrus.Entry
}

func NewOrderService(repo domain.OrderRepository, cache domain.Cache, logger *logrus.Entry) *OrderService {
	return &OrderService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *OrderService) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	log := s.logger.WithField("uid", order.OrderUID)
	log.Info("processing new order")

	if err := s.repo.Save(ctx, order); err != nil {
		log.Errorf("failed to save order: %v", err)
		return err
	}
	log.Info("saved order successfully")
	return nil
}

func (s *OrderService) GetOrderByUID(ctx context.Context, uid uuid.UUID) (*domain.Order, error) {
	log := s.logger.WithField("order_uid", uid)
	log.Info("getting order by uid")

	order, err := s.repo.GetByUID(ctx, uid)
	if err != nil {
		log.Warnf("failed to get order by uid: %v", err)
		return nil, err
	}

	log.Info("order successfully retrieved")
	return order, nil
}

func (s *OrderService) WarmUpCache(ctx context.Context, limit uint) error {
	s.logger.Info("starting cache warm-up")

	orders, err := s.repo.GetLatest(ctx, limit)
	if err != nil {
		s.logger.Errorf("failed to get latest orders for cache warm-up: %v", err)
		return err
	}

	for _, order := range orders {
		_ = s.cache.Set(ctx, order)
	}

	s.logger.Infof("cache warm-up finished. %d orders were cached.", len(orders))
	return nil
}
