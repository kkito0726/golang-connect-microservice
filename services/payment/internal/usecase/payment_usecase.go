package usecase

import (
	"context"
	"fmt"
	"log/slog"

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
		// 決済は成功済みのため失敗を返さず、オペレーターが手動で整合性を確認する
		slog.Error("payment created but order status update failed - manual reconciliation required",
			"order_id", orderID, "payment_id", payment.ID, "error", err)
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
		// 払い戻しは成功済みのため失敗を返さず、オペレーターが手動で整合性を確認する
		slog.Error("payment refunded but order cancellation failed - manual reconciliation required",
			"order_id", payment.OrderID, "payment_id", payment.ID, "error", err)
	}

	return payment, nil
}
