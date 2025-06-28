package repository

import (
	"L0/internal/domain"
	"context"
	"github.com/google/uuid"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error

	GetByUID(ctx context.Context, uid uuid.UUID) (*domain.Order, error)
	Get
}
