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

	db "github.com/ken/connect-microservice/services/payment/db/sqlc"
	"github.com/ken/connect-microservice/services/payment/internal/domain"
)

type PaymentRepository struct {
	queries *db.Queries
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{queries: db.New(pool)}
}

var _ domain.PaymentRepository = (*PaymentRepository)(nil)

func (r *PaymentRepository) Create(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	row, err := r.queries.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:     p.OrderID,
		UserID:      p.UserID,
		AmountCents: p.AmountCents,
		Status:      p.Status,
		Method:      p.Method,
	})
	if err != nil {
		return domain.Payment{}, fmt.Errorf("insert payment: %w", err)
	}
	return toPayment(row), nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (domain.Payment, error) {
	row, err := r.queries.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Payment{}, fmt.Errorf("get payment %s: %w", id, domain.ErrNotFound)
		}
		return domain.Payment{}, fmt.Errorf("get payment: %w", err)
	}
	return toPayment(row), nil
}

func (r *PaymentRepository) List(ctx context.Context, orderID, userID string, limit, offset int) ([]domain.Payment, int, error) {
	total, err := r.queries.CountPayments(ctx, db.CountPaymentsParams{
		OrderID: stringToUUID(orderID),
		UserID:  stringToUUID(userID),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	rows, err := r.queries.ListPayments(ctx, db.ListPaymentsParams{
		OrderID: stringToUUID(orderID),
		UserID:  stringToUUID(userID),
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}

	payments := make([]domain.Payment, len(rows))
	for i, row := range rows {
		payments[i] = toPayment(row)
	}
	return payments, int(total), nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status string) (domain.Payment, error) {
	row, err := r.queries.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		Status: status,
		ID:     id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Payment{}, fmt.Errorf("update payment status %s: %w", id, domain.ErrNotFound)
		}
		return domain.Payment{}, fmt.Errorf("update payment status: %w", err)
	}
	return toPayment(row), nil
}

func toPayment(p db.Payment) domain.Payment {
	return domain.Payment{
		ID:          p.ID,
		OrderID:     p.OrderID,
		UserID:      p.UserID,
		AmountCents: p.AmountCents,
		Status:      p.Status,
		Method:      p.Method,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
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
