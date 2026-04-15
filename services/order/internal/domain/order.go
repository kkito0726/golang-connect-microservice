package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrInsufficientStock = errors.New("insufficient stock")
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
	Create(ctx context.Context, userID string, items []OrderItem, totalCents int64) (Order, error)
	GetByID(ctx context.Context, id string) (Order, error)
	List(ctx context.Context, userID, status string, limit, offset int) ([]Order, int, error)
	UpdateStatus(ctx context.Context, id, status string) (Order, error)
}
