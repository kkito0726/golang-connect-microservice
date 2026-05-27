package repository

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/ken/connect-microservice/services/order/db/sqlc"
	"github.com/ken/connect-microservice/services/order/internal/domain"
)

type OrderRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool, queries: db.New(pool)}
}

var _ domain.OrderRepository = (*OrderRepository)(nil)

func (r *OrderRepository) Create(ctx context.Context, in domain.Order) (domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	orderRow, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		UserID:     in.UserID,
		Status:     in.Status,
		TotalCents: in.TotalCents,
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}

	items := make([]domain.OrderItem, 0, len(in.Items))
	for _, item := range in.Items {
		itemRow, err := qtx.CreateOrderItem(ctx, db.CreateOrderItemParams{
			OrderID:        orderRow.ID,
			ProductID:      item.ProductID,
			ProductName:    item.ProductName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
		})
		if err != nil {
			return domain.Order{}, fmt.Errorf("insert order item: %w", err)
		}
		items = append(items, toOrderItem(itemRow))
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("commit transaction: %w", err)
	}

	return domain.Order{
		ID:         orderRow.ID,
		UserID:     orderRow.UserID,
		Status:     orderRow.Status,
		TotalCents: orderRow.TotalCents,
		Items:      items,
		CreatedAt:  orderRow.CreatedAt.Time,
		UpdatedAt:  orderRow.UpdatedAt.Time,
	}, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (domain.Order, error) {
	orderRow, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, fmt.Errorf("get order %s: %w", id, domain.ErrNotFound)
		}
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}

	itemRows, err := r.queries.GetOrderItems(ctx, id)
	if err != nil {
		return domain.Order{}, fmt.Errorf("list order items: %w", err)
	}

	items := make([]domain.OrderItem, len(itemRows))
	for i, row := range itemRows {
		items[i] = toOrderItem(row)
	}

	return domain.Order{
		ID:         orderRow.ID,
		UserID:     orderRow.UserID,
		Status:     orderRow.Status,
		TotalCents: orderRow.TotalCents,
		Items:      items,
		CreatedAt:  orderRow.CreatedAt.Time,
		UpdatedAt:  orderRow.UpdatedAt.Time,
	}, nil
}

func (r *OrderRepository) List(ctx context.Context, userID, status string, limit, offset int) ([]domain.Order, int, error) {
	total, err := r.queries.CountOrders(ctx, db.CountOrdersParams{
		UserID: stringToUUID(userID),
		Status: nullableString(status),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	orderRows, err := r.queries.ListOrders(ctx, db.ListOrdersParams{
		UserID: stringToUUID(userID),
		Status: nullableString(status),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}

	orders := make([]domain.Order, 0, len(orderRows))
	for _, orderRow := range orderRows {
		itemRows, err := r.queries.GetOrderItems(ctx, orderRow.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("list order items: %w", err)
		}
		items := make([]domain.OrderItem, len(itemRows))
		for i, row := range itemRows {
			items[i] = toOrderItem(row)
		}
		orders = append(orders, domain.Order{
			ID:         orderRow.ID,
			UserID:     orderRow.UserID,
			Status:     orderRow.Status,
			TotalCents: orderRow.TotalCents,
			Items:      items,
			CreatedAt:  orderRow.CreatedAt.Time,
			UpdatedAt:  orderRow.UpdatedAt.Time,
		})
	}
	return orders, int(total), nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id, status string) (domain.Order, error) {
	orderRow, err := r.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		Status: status,
		ID:     id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, fmt.Errorf("update order status %s: %w", id, domain.ErrNotFound)
		}
		return domain.Order{}, fmt.Errorf("update order status: %w", err)
	}

	itemRows, err := r.queries.GetOrderItems(ctx, id)
	if err != nil {
		return domain.Order{}, fmt.Errorf("list order items: %w", err)
	}

	items := make([]domain.OrderItem, len(itemRows))
	for i, row := range itemRows {
		items[i] = toOrderItem(row)
	}

	return domain.Order{
		ID:         orderRow.ID,
		UserID:     orderRow.UserID,
		Status:     orderRow.Status,
		TotalCents: orderRow.TotalCents,
		Items:      items,
		CreatedAt:  orderRow.CreatedAt.Time,
		UpdatedAt:  orderRow.UpdatedAt.Time,
	}, nil
}

func toOrderItem(row db.OrderItem) domain.OrderItem {
	return domain.OrderItem{
		ID:             row.ID,
		OrderID:        row.OrderID,
		ProductID:      row.ProductID,
		ProductName:    row.ProductName,
		Quantity:       row.Quantity,
		UnitPriceCents: row.UnitPriceCents,
	}
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringToUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{}
	}
	cleaned := strings.ReplaceAll(s, "-", "")
	if len(cleaned) != 32 {
		return pgtype.UUID{}
	}
	var b [16]byte
	if _, err := hex.Decode(b[:], []byte(cleaned)); err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: b, Valid: true}
}
