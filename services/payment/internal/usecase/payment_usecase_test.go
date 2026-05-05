package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ken/connect-microservice/services/payment/internal/domain"
)

// ---- モック ----------------------------------------------------------------

type mockPaymentRepo struct {
	payments map[string]domain.Payment
	createFn func(ctx context.Context, p domain.Payment) (domain.Payment, error)
}

func (m *mockPaymentRepo) Create(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	p.ID = "payment-1"
	return p, nil
}

func (m *mockPaymentRepo) GetByID(_ context.Context, id string) (domain.Payment, error) {
	p, ok := m.payments[id]
	if !ok {
		return domain.Payment{}, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockPaymentRepo) List(_ context.Context, _, _ string, _, _ int) ([]domain.Payment, int, error) {
	return nil, 0, nil
}

func (m *mockPaymentRepo) UpdateStatus(_ context.Context, id, status string) (domain.Payment, error) {
	p, ok := m.payments[id]
	if !ok {
		return domain.Payment{}, domain.ErrNotFound
	}
	p.Status = status
	return p, nil
}

type mockOrderClient struct {
	orders           map[string]domain.OrderInfo
	updateStatusErr  error
	cancelErr        error
	updateStatusCalled bool
	cancelCalled       bool
}

func (m *mockOrderClient) GetOrder(_ context.Context, id string) (domain.OrderInfo, error) {
	o, ok := m.orders[id]
	if !ok {
		return domain.OrderInfo{}, domain.ErrNotFound
	}
	return o, nil
}

func (m *mockOrderClient) UpdateOrderStatus(_ context.Context, _, _ string) error {
	m.updateStatusCalled = true
	return m.updateStatusErr
}

func (m *mockOrderClient) CancelOrder(_ context.Context, _ string) error {
	m.cancelCalled = true
	return m.cancelErr
}

// ---- CreatePayment テスト --------------------------------------------------

// 正常系: 決済が作成され、注文ステータスも更新される
func TestCreatePayment_Success(t *testing.T) {
	t.Parallel()

	orderClient := &mockOrderClient{
		orders: map[string]domain.OrderInfo{
			"order-1": {ID: "order-1", UserID: "u1", TotalCents: 5000},
		},
	}
	uc := NewPaymentUsecase(&mockPaymentRepo{}, orderClient)

	payment, err := uc.CreatePayment(context.Background(), "order-1", "u1", "credit_card")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if payment.AmountCents != 5000 {
		t.Errorf("AmountCents = %d, want 5000", payment.AmountCents)
	}
	if !orderClient.updateStatusCalled {
		t.Error("UpdateOrderStatus should be called")
	}
}

// UpdateOrderStatusが失敗しても決済は返る（手動整合性確認フロー）
func TestCreatePayment_UpdateOrderStatusFails_ReturnsPaymentAnyway(t *testing.T) {
	t.Parallel()

	orderClient := &mockOrderClient{
		orders: map[string]domain.OrderInfo{
			"order-1": {ID: "order-1", UserID: "u1", TotalCents: 5000},
		},
		updateStatusErr: errors.New("order service unavailable"),
	}
	uc := NewPaymentUsecase(&mockPaymentRepo{}, orderClient)

	payment, err := uc.CreatePayment(context.Background(), "order-1", "u1", "credit_card")

	if err != nil {
		t.Fatalf("want no error even when UpdateOrderStatus fails, got %v", err)
	}
	if payment.ID == "" {
		t.Error("payment should be returned even when order status update fails")
	}
	if payment.Status != "completed" {
		t.Errorf("payment.Status = %q, want %q", payment.Status, "completed")
	}
}

// 別ユーザーの注文には決済できない
func TestCreatePayment_WrongUser_ReturnsError(t *testing.T) {
	t.Parallel()

	orderClient := &mockOrderClient{
		orders: map[string]domain.OrderInfo{
			"order-1": {ID: "order-1", UserID: "u1", TotalCents: 5000},
		},
	}
	uc := NewPaymentUsecase(&mockPaymentRepo{}, orderClient)

	_, err := uc.CreatePayment(context.Background(), "order-1", "other-user", "credit_card")

	if err == nil {
		t.Fatal("want error for wrong user, got nil")
	}
}

// ---- RefundPayment テスト --------------------------------------------------

// 正常系: 払い戻しが完了し、注文もキャンセルされる
func TestRefundPayment_Success(t *testing.T) {
	t.Parallel()

	orderClient := &mockOrderClient{}
	repo := &mockPaymentRepo{
		payments: map[string]domain.Payment{
			"pay-1": {ID: "pay-1", OrderID: "order-1", Status: "completed"},
		},
	}
	uc := NewPaymentUsecase(repo, orderClient)

	payment, err := uc.RefundPayment(context.Background(), "pay-1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if payment.Status != "refunded" {
		t.Errorf("payment.Status = %q, want %q", payment.Status, "refunded")
	}
	if !orderClient.cancelCalled {
		t.Error("CancelOrder should be called")
	}
}

// CancelOrderが失敗しても払い戻し済みの決済は返る（手動整合性確認フロー）
func TestRefundPayment_CancelOrderFails_ReturnsPaymentAnyway(t *testing.T) {
	t.Parallel()

	orderClient := &mockOrderClient{
		cancelErr: errors.New("order service unavailable"),
	}
	repo := &mockPaymentRepo{
		payments: map[string]domain.Payment{
			"pay-1": {ID: "pay-1", OrderID: "order-1", Status: "completed"},
		},
	}
	uc := NewPaymentUsecase(repo, orderClient)

	payment, err := uc.RefundPayment(context.Background(), "pay-1")

	if err != nil {
		t.Fatalf("want no error even when CancelOrder fails, got %v", err)
	}
	if payment.Status != "refunded" {
		t.Errorf("payment.Status = %q, want %q", payment.Status, "refunded")
	}
}

// completed 以外の決済は払い戻せない
func TestRefundPayment_NotCompleted_ReturnsError(t *testing.T) {
	t.Parallel()

	repo := &mockPaymentRepo{
		payments: map[string]domain.Payment{
			"pay-1": {ID: "pay-1", OrderID: "order-1", Status: "refunded"},
		},
	}
	uc := NewPaymentUsecase(repo, &mockOrderClient{})

	_, err := uc.RefundPayment(context.Background(), "pay-1")

	if err == nil {
		t.Fatal("want error for non-completed payment, got nil")
	}
}
