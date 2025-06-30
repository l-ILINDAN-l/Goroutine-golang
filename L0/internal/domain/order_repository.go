package domain

import (
	"context"
	"github.com/google/uuid"
)

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	GetByUID(ctx context.Context, uid uuid.UUID) (*Order, error)
	GetLatest(ctx context.Context, limit uint) ([]*Order, error)
}
