package usecase

import (
	"context"
	"fmt"

	"github.com/ken/connect-microservice/services/payment/internal/repository"
)

// OrderInfo はユースケースが外部サービスから必要とする注文情報。
type OrderInfo struct {
	ID         string
	UserID     string
	TotalCents int64
}

// OrderClient は注文サービスへの依存を抽象化するインターフェース。
type OrderClient interface {
	GetOrder(ctx context.Context, id string) (OrderInfo, error)
	UpdateOrderStatus(ctx context.Context, id, status string) error
	CancelOrder(ctx context.Context, id string) error
}

// PaymentRepository はデータ永続化への依存を抽象化するインターフェース。
type PaymentRepository interface {
	Create(ctx context.Context, p repository.Payment) (repository.Payment, error)
	GetByID(ctx context.Context, id string) (repository.Payment, error)
	List(ctx context.Context, orderID, userID string, limit, offset int) ([]repository.Payment, int, error)
	UpdateStatus(ctx context.Context, id, status string) (repository.Payment, error)
}

type PaymentUsecase struct {
	repo        PaymentRepository
	orderClient OrderClient
}

func NewPaymentUsecase(repo PaymentRepository, orderClient OrderClient) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, orderClient: orderClient}
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, orderID, userID, method string) (repository.Payment, error) {
	order, err := uc.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		return repository.Payment{}, fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return repository.Payment{}, fmt.Errorf("order does not belong to user")
	}

	payment, err := uc.repo.Create(ctx, repository.Payment{
		OrderID:     orderID,
		UserID:      userID,
		AmountCents: order.TotalCents,
		Status:      "completed",
		Method:      method,
	})
	if err != nil {
		return repository.Payment{}, fmt.Errorf("create payment: %w", err)
	}

	if err := uc.orderClient.UpdateOrderStatus(ctx, orderID, "confirmed"); err != nil {
		return repository.Payment{}, fmt.Errorf("update order status: %w", err)
	}

	return payment, nil
}

func (uc *PaymentUsecase) GetPayment(ctx context.Context, id string) (repository.Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *PaymentUsecase) ListPayments(ctx context.Context, orderID, userID string, pageSize, page int) ([]repository.Payment, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return uc.repo.List(ctx, orderID, userID, pageSize, (page-1)*pageSize)
}

func (uc *PaymentUsecase) RefundPayment(ctx context.Context, id string) (repository.Payment, error) {
	payment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return repository.Payment{}, err
	}
	if payment.Status != "completed" {
		return repository.Payment{}, fmt.Errorf("can only refund completed payments, current status: %s", payment.Status)
	}

	payment, err = uc.repo.UpdateStatus(ctx, id, "refunded")
	if err != nil {
		return repository.Payment{}, err
	}

	if err := uc.orderClient.CancelOrder(ctx, payment.OrderID); err != nil {
		return repository.Payment{}, fmt.Errorf("cancel order: %w", err)
	}

	return payment, nil
}
