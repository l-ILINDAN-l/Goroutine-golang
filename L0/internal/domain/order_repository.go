package domain

import (
	"context"
)

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	GetByUID(ctx context.Context, uid string) (*Order, error)
	GetLatest(ctx context.Context, limit uint) ([]*Order, error)
}
