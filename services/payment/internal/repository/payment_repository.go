package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p Payment) (Payment, error) {
	var result Payment
	err := r.pool.QueryRow(ctx,
		`INSERT INTO payments (order_id, user_id, amount_cents, status, method)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, order_id, user_id, amount_cents, status, method, created_at, updated_at`,
		p.OrderID, p.UserID, p.AmountCents, p.Status, p.Method,
	).Scan(&result.ID, &result.OrderID, &result.UserID, &result.AmountCents,
		&result.Status, &result.Method, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		return Payment{}, fmt.Errorf("insert payment: %w", err)
	}
	return result, nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (Payment, error) {
	var p Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_id, user_id, amount_cents, status, method, created_at, updated_at
		 FROM payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.AmountCents, &p.Status, &p.Method, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Payment{}, fmt.Errorf("payment not found: %s", id)
		}
		return Payment{}, fmt.Errorf("get payment: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) List(ctx context.Context, orderID, userID string, limit, offset int) ([]Payment, int, error) {
	query := `SELECT id, order_id, user_id, amount_cents, status, method, created_at, updated_at FROM payments WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM payments WHERE 1=1`
	args := []any{}
	argIdx := 1

	if orderID != "" {
		query += fmt.Sprintf(` AND order_id = $%d`, argIdx)
		countQuery += fmt.Sprintf(` AND order_id = $%d`, argIdx)
		args = append(args, orderID)
		argIdx++
	}
	if userID != "" {
		query += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		countQuery += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		args = append(args, userID)
		argIdx++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.AmountCents, &p.Status, &p.Method, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, total, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status string) (Payment, error) {
	var p Payment
	err := r.pool.QueryRow(ctx,
		`UPDATE payments SET status = $1, updated_at = now()
		 WHERE id = $2
		 RETURNING id, order_id, user_id, amount_cents, status, method, created_at, updated_at`,
		status, id,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.AmountCents, &p.Status, &p.Method, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Payment{}, fmt.Errorf("payment not found: %s", id)
		}
		return Payment{}, fmt.Errorf("update payment status: %w", err)
	}
	return p, nil
}
