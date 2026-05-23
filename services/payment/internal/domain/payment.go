package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidStatus = errors.New("invalid status transition")
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

// NewPayment builds a Payment aggregate ready for persistence.
func NewPayment(orderID, userID, method string, amountCents int64) Payment {
	return Payment{
		OrderID:     orderID,
		UserID:      userID,
		AmountCents: amountCents,
		Status:      "completed",
		Method:      method,
	}
}

// Refund validates the status transition and returns a new refunded Payment.
func (p Payment) Refund() (Payment, error) {
	if p.Status != "completed" {
		return Payment{}, fmt.Errorf("%w: cannot refund payment with status %q", ErrInvalidStatus, p.Status)
	}
	refunded := p
	refunded.Status = "refunded"
	return refunded, nil
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
