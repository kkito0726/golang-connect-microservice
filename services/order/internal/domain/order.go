package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidStatus     = errors.New("invalid status transition")
)

type Order struct {
	ID         string
	UserID     string
	Status     string
	TotalCents int64
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewOrder builds a valid Order aggregate, validating stock availability and computing the total.
func NewOrder(userID string, items []CreateOrderItem, products map[string]ProductInfo) (Order, error) {
	var orderItems []OrderItem
	var totalCents int64

	for _, item := range items {
		product, ok := products[item.ProductID]
		if !ok {
			return Order{}, fmt.Errorf("product %s: %w", item.ProductID, ErrNotFound)
		}
		if product.StockQuantity < item.Quantity {
			return Order{}, fmt.Errorf("product %q: %w", product.Name, ErrInsufficientStock)
		}
		orderItems = append(orderItems, OrderItem{
			ProductID:      item.ProductID,
			ProductName:    product.Name,
			Quantity:       item.Quantity,
			UnitPriceCents: product.PriceCents,
		})
		totalCents += product.PriceCents * int64(item.Quantity)
	}

	return Order{
		UserID:     userID,
		Status:     "pending",
		TotalCents: totalCents,
		Items:      orderItems,
	}, nil
}

// Cancel validates the status transition and returns a new cancelled Order.
func (o Order) Cancel() (Order, error) {
	if o.Status != "pending" {
		return Order{}, fmt.Errorf("%w: cannot cancel order with status %q", ErrInvalidStatus, o.Status)
	}
	cancelled := o
	cancelled.Status = "cancelled"
	return cancelled, nil
}

type OrderItem struct {
	ID             string
	OrderID        string
	ProductID      string
	ProductName    string
	Quantity       int32
	UnitPriceCents int64
}

type ProductInfo struct {
	ID            string
	Name          string
	PriceCents    int64
	StockQuantity int32
}

type CreateOrderItem struct {
	ProductID string
	Quantity  int32
}

type CreateOrderInput struct {
	UserID string
	Items  []CreateOrderItem
}

type UserClient interface {
	ValidateUser(ctx context.Context, id string) error
}

type ProductClient interface {
	GetProduct(ctx context.Context, id string) (ProductInfo, error)
	DeductStock(ctx context.Context, productID string, quantity int32) error
	RestoreStock(ctx context.Context, productID string, quantity int32, referenceID string) error
}

type OrderRepository interface {
	Create(ctx context.Context, order Order) (Order, error)
	GetByID(ctx context.Context, id string) (Order, error)
	List(ctx context.Context, userID, status string, limit, offset int) ([]Order, int, error)
	UpdateStatus(ctx context.Context, id, status string) (Order, error)
}
