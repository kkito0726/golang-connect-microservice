package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type Payment struct {
	ID          string
	OrderID     string
	UserID      string
	AmountCents int64
	Status      string
	Method      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderInfo struct {
	ID         string
	UserID     string
	TotalCents int64
}

type OrderClient interface {
	GetOrder(ctx context.Context, id string) (OrderInfo, error)
	UpdateOrderStatus(ctx context.Context, id, status string) error
	CancelOrder(ctx context.Context, id string) error
}

type PaymentRepository interface {
	Create(ctx context.Context, p Payment) (Payment, error)
	GetByID(ctx context.Context, id string) (Payment, error)
	List(ctx context.Context, orderID, userID string, limit, offset int) ([]Payment, int, error)
	UpdateStatus(ctx context.Context, id, status string) (Payment, error)
}
