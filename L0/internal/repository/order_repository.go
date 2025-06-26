package repository

import (
	"L0/internal/domain"
	"context"
	"github.com/google/uuid"
)

type OrderRepository interface {
	GetByUID(ctx context.Context, uid uuid.UUID) (*domain.Order, error)
	Save(ctx context.Context, order *domain.Order) error
}
