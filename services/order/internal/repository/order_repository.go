package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ken/connect-microservice/services/order/internal/domain"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

var _ domain.OrderRepository = (*OrderRepository)(nil)

func (r *OrderRepository) Create(ctx context.Context, userID string, items []domain.OrderItem, totalCents int64) (domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var order domain.Order
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, status, total_cents)
		 VALUES ($1, 'pending', $2)
		 RETURNING id, user_id, status, total_cents, created_at, updated_at`,
		userID, totalCents,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalCents, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}

	for _, item := range items {
		var oi domain.OrderItem
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price_cents)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, order_id, product_id, product_name, quantity, unit_price_cents`,
			order.ID, item.ProductID, item.ProductName, item.Quantity, item.UnitPriceCents,
		).Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.ProductName, &oi.Quantity, &oi.UnitPriceCents)
		if err != nil {
			return domain.Order{}, fmt.Errorf("insert order item: %w", err)
		}
		order.Items = append(order.Items, oi)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("commit transaction: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (domain.Order, error) {
	var order domain.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total_cents, created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalCents, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Order{}, fmt.Errorf("order not found: %s", id)
		}
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}

	items, err := r.getOrderItems(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	order.Items = items
	return order, nil
}

func (r *OrderRepository) getOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, product_id, product_name, quantity, unit_price_cents
		 FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.Quantity, &item.UnitPriceCents); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *OrderRepository) List(ctx context.Context, userID, status string, limit, offset int) ([]domain.Order, int, error) {
	query := `SELECT id, user_id, status, total_cents, created_at, updated_at FROM orders WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM orders WHERE 1=1`
	args := []any{}
	argIdx := 1

	if userID != "" {
		query += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		countQuery += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		args = append(args, userID)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(` AND status = $%d`, argIdx)
		countQuery += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalCents, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		items, err := r.getOrderItems(ctx, o.ID)
		if err != nil {
			return nil, 0, err
		}
		o.Items = items
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id, status string) (domain.Order, error) {
	var order domain.Order
	err := r.pool.QueryRow(ctx,
		`UPDATE orders SET status = $1, updated_at = now()
		 WHERE id = $2
		 RETURNING id, user_id, status, total_cents, created_at, updated_at`,
		status, id,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalCents, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Order{}, fmt.Errorf("order not found: %s", id)
		}
		return domain.Order{}, fmt.Errorf("update order status: %w", err)
	}

	items, err := r.getOrderItems(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	order.Items = items
	return order, nil
}
