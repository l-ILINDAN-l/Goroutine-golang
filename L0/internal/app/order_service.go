package app

import (
	"L0/internal/domain"
	"L0/internal/metrics"
	"context"
	"github.com/sirupsen/logrus"
)

type OrderService interface {
	ProcessNewOrder(ctx context.Context, order *domain.Order) error
	GetOrderByUID(ctx context.Context, uid string) (*domain.Order, error)
	WarmUpCache(ctx context.Context, limit uint) error
}

type orderService struct {
	repo   domain.OrderRepository
	cache  domain.Cache
	logger *logrus.Entry
}

func NewOrderService(repo domain.OrderRepository, cache domain.Cache, logger *logrus.Entry) OrderService {
	return &orderService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *orderService) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	log := s.logger.WithField("uid", order.OrderUID)
	log.Info("processing new order")

	if err := s.repo.Save(ctx, order); err != nil {
		log.Errorf("failed to save order: %v", err)
		return err
	}
	log.Info("saved order successfully")
	metrics.OrdersProcessedTotal.Inc()
	return nil
}

func (s *orderService) GetOrderByUID(ctx context.Context, uid string) (*domain.Order, error) {
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

func (s *orderService) WarmUpCache(ctx context.Context, limit uint) error {
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
