package domain

import (
	"context"
	"github.com/google/uuid"
)

type Cache interface {
	Set(ctx context.Context, order *Order) error
	Get(ctx context.Context, uid uuid.UUID) (*Order, bool, error)
}
