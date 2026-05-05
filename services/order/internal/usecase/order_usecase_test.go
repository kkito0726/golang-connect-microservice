package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ken/connect-microservice/services/order/internal/domain"
)

// ---- モック ----------------------------------------------------------------

type mockUserClient struct {
	validateErr error
}

func (m *mockUserClient) ValidateUser(_ context.Context, _ string) error {
	return m.validateErr
}

type deductCall struct {
	productID string
	quantity  int32
}

type restoreCall struct {
	productID string
	quantity  int32
}

type mockProductClient struct {
	products    map[string]domain.ProductInfo
	deductErrs  map[string]error // productID → error（nilなら成功）
	deducted    []deductCall
	restored    []restoreCall
	restoreErrs map[string]error
}

func (m *mockProductClient) GetProduct(_ context.Context, id string) (domain.ProductInfo, error) {
	p, ok := m.products[id]
	if !ok {
		return domain.ProductInfo{}, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockProductClient) DeductStock(_ context.Context, productID string, quantity int32) error {
	m.deducted = append(m.deducted, deductCall{productID, quantity})
	if err, ok := m.deductErrs[productID]; ok {
		return err
	}
	return nil
}

func (m *mockProductClient) RestoreStock(_ context.Context, productID string, quantity int32, _ string) error {
	m.restored = append(m.restored, restoreCall{productID, quantity})
	if err, ok := m.restoreErrs[productID]; ok {
		return err
	}
	return nil
}

type mockOrderRepo struct {
	createFn func(ctx context.Context, userID string, items []domain.OrderItem, totalCents int64) (domain.Order, error)
}

func (m *mockOrderRepo) Create(ctx context.Context, userID string, items []domain.OrderItem, totalCents int64) (domain.Order, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, items, totalCents)
	}
	return domain.Order{ID: "order-1", UserID: userID, Items: items, TotalCents: totalCents}, nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id string) (domain.Order, error) {
	return domain.Order{ID: id}, nil
}

func (m *mockOrderRepo) List(_ context.Context, _, _ string, _, _ int) ([]domain.Order, int, error) {
	return nil, 0, nil
}

func (m *mockOrderRepo) UpdateStatus(_ context.Context, id, status string) (domain.Order, error) {
	return domain.Order{ID: id, Status: status}, nil
}

// ---- テスト ----------------------------------------------------------------

// 正常系: 全商品のDeductStockが成功し、注文が作成される
func TestCreateOrder_Success(t *testing.T) {
	t.Parallel()

	productClient := &mockProductClient{
		products: map[string]domain.ProductInfo{
			"p1": {ID: "p1", Name: "Product1", PriceCents: 1000, StockQuantity: 10},
			"p2": {ID: "p2", Name: "Product2", PriceCents: 2000, StockQuantity: 5},
		},
	}
	uc := NewOrderUsecase(&mockOrderRepo{}, productClient, &mockUserClient{})

	order, err := uc.CreateOrder(context.Background(), domain.CreateOrderInput{
		UserID: "u1",
		Items: []domain.CreateOrderItem{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
		},
	})

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if order.TotalCents != 4000 {
		t.Errorf("TotalCents = %d, want 4000", order.TotalCents)
	}
	if len(productClient.deducted) != 2 {
		t.Errorf("DeductStock called %d times, want 2", len(productClient.deducted))
	}
	if len(productClient.restored) != 0 {
		t.Errorf("RestoreStock should not be called on success, called %d times", len(productClient.restored))
	}
}

// Saga補償: 2番目のDeductStockが失敗したとき、1番目の分がRestoreされる
func TestCreateOrder_SagaCompensation_RestoresDeductedStock(t *testing.T) {
	t.Parallel()

	productClient := &mockProductClient{
		products: map[string]domain.ProductInfo{
			"p1": {ID: "p1", Name: "Product1", PriceCents: 1000, StockQuantity: 10},
			"p2": {ID: "p2", Name: "Product2", PriceCents: 2000, StockQuantity: 5},
		},
		deductErrs: map[string]error{
			"p2": errors.New("deduct failed"),
		},
	}
	uc := NewOrderUsecase(&mockOrderRepo{}, productClient, &mockUserClient{})

	_, err := uc.CreateOrder(context.Background(), domain.CreateOrderInput{
		UserID: "u1",
		Items: []domain.CreateOrderItem{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
		},
	})

	if err == nil {
		t.Fatal("want error, got nil")
	}

	// p1の分がRestoreされていること
	if len(productClient.restored) != 1 {
		t.Fatalf("RestoreStock called %d times, want 1", len(productClient.restored))
	}
	if productClient.restored[0].productID != "p1" {
		t.Errorf("restored product = %q, want %q", productClient.restored[0].productID, "p1")
	}
	if productClient.restored[0].quantity != 2 {
		t.Errorf("restored quantity = %d, want 2", productClient.restored[0].quantity)
	}
}

// Saga補償: 1番目のDeductStockが失敗したとき、RestoreStockは呼ばれない
func TestCreateOrder_SagaCompensation_NoRestoreWhenFirstFails(t *testing.T) {
	t.Parallel()

	productClient := &mockProductClient{
		products: map[string]domain.ProductInfo{
			"p1": {ID: "p1", Name: "Product1", PriceCents: 1000, StockQuantity: 10},
			"p2": {ID: "p2", Name: "Product2", PriceCents: 2000, StockQuantity: 5},
		},
		deductErrs: map[string]error{
			"p1": errors.New("deduct failed"),
		},
	}
	uc := NewOrderUsecase(&mockOrderRepo{}, productClient, &mockUserClient{})

	_, err := uc.CreateOrder(context.Background(), domain.CreateOrderInput{
		UserID: "u1",
		Items: []domain.CreateOrderItem{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
		},
	})

	if err == nil {
		t.Fatal("want error, got nil")
	}
	if len(productClient.restored) != 0 {
		t.Errorf("RestoreStock called %d times, want 0 (nothing was deducted yet)", len(productClient.restored))
	}
}

// Saga補償: RestoreStock自体が失敗してもCreateOrderはエラーを返す（補償失敗はログのみ）
func TestCreateOrder_SagaCompensation_CompensationFailureDoesNotPanic(t *testing.T) {
	t.Parallel()

	productClient := &mockProductClient{
		products: map[string]domain.ProductInfo{
			"p1": {ID: "p1", Name: "Product1", PriceCents: 1000, StockQuantity: 10},
			"p2": {ID: "p2", Name: "Product2", PriceCents: 2000, StockQuantity: 5},
		},
		deductErrs: map[string]error{
			"p2": errors.New("deduct failed"),
		},
		restoreErrs: map[string]error{
			"p1": errors.New("restore also failed"),
		},
	}
	uc := NewOrderUsecase(&mockOrderRepo{}, productClient, &mockUserClient{})

	_, err := uc.CreateOrder(context.Background(), domain.CreateOrderInput{
		UserID: "u1",
		Items: []domain.CreateOrderItem{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
		},
	})

	if err == nil {
		t.Fatal("want error, got nil")
	}
	// RestoreStockが呼ばれていること（失敗してもパニックしないことを確認）
	if len(productClient.restored) != 1 {
		t.Errorf("RestoreStock called %d times, want 1", len(productClient.restored))
	}
}
