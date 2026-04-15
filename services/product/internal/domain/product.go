package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Product struct {
	ID            string
	SKU           string
	Name          string
	Description   string
	PriceCents    int64
	StockQuantity int32
	Category      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type StockMovement struct {
	ID          string
	ProductID   string
	Delta       int32
	Reason      string
	ReferenceID string
	CreatedAt   time.Time
}

type ProductRepository interface {
	Create(ctx context.Context, p Product) (Product, error)
	GetByID(ctx context.Context, id string) (Product, error)
	List(ctx context.Context, limit, offset int, category string) ([]Product, int, error)
	Update(ctx context.Context, id, name, description, category string, priceCents int64) (Product, error)
	SoftDelete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, productID string, delta int32, reason, referenceID string) (Product, StockMovement, error)
	GetStockMovements(ctx context.Context, productID string, limit int) ([]StockMovement, error)
}
