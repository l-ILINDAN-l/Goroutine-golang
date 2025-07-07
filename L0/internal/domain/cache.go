package domain

import (
	"context"
)

type Cache interface {
	Set(ctx context.Context, order *Order) error
	Get(ctx context.Context, uid string) (*Order, bool, error)
}
