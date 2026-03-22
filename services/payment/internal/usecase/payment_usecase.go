package usecase

import (
	"context"
	"fmt"

	"github.com/ken/connect-microservice/services/payment/internal/domain"
)

type PaymentUsecase struct {
	repo        domain.PaymentRepository
	orderClient domain.OrderClient
}

func NewPaymentUsecase(repo domain.PaymentRepository, orderClient domain.OrderClient) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, orderClient: orderClient}
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, orderID, userID, method string) (domain.Payment, error) {
	order, err := uc.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return domain.Payment{}, fmt.Errorf("order does not belong to user")
	}

	payment, err := uc.repo.Create(ctx, domain.Payment{
		OrderID:     orderID,
		UserID:      userID,
		AmountCents: order.TotalCents,
		Status:      "completed",
		Method:      method,
	})
	if err != nil {
		return domain.Payment{}, fmt.Errorf("create payment: %w", err)
	}

	if err := uc.orderClient.UpdateOrderStatus(ctx, orderID, "confirmed"); err != nil {
		return domain.Payment{}, fmt.Errorf("update order status: %w", err)
	}

	return payment, nil
}

func (uc *PaymentUsecase) GetPayment(ctx context.Context, id string) (domain.Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *PaymentUsecase) ListPayments(ctx context.Context, orderID, userID string, pageSize, page int) ([]domain.Payment, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return uc.repo.List(ctx, orderID, userID, pageSize, (page-1)*pageSize)
}

func (uc *PaymentUsecase) RefundPayment(ctx context.Context, id string) (domain.Payment, error) {
	payment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Payment{}, err
	}
	if payment.Status != "completed" {
		return domain.Payment{}, fmt.Errorf("can only refund completed payments, current status: %s", payment.Status)
	}

	payment, err = uc.repo.UpdateStatus(ctx, id, "refunded")
	if err != nil {
		return domain.Payment{}, err
	}

	if err := uc.orderClient.CancelOrder(ctx, payment.OrderID); err != nil {
		return domain.Payment{}, fmt.Errorf("cancel order: %w", err)
	}

	return payment, nil
}
