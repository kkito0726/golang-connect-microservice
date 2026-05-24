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
	createFn       func(ctx context.Context, order domain.Order) (domain.Order, error)
	orders         map[string]domain.Order
	lastListLimit  int
	lastListOffset int
}

func (m *mockOrderRepo) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	if m.createFn != nil {
		return m.createFn(ctx, order)
	}
	return domain.Order{ID: "order-1", UserID: order.UserID, Items: order.Items, TotalCents: order.TotalCents, Status: order.Status}, nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id string) (domain.Order, error) {
	if m.orders != nil {
		if o, ok := m.orders[id]; ok {
			return o, nil
		}
		return domain.Order{}, domain.ErrNotFound
	}
	return domain.Order{ID: id, Status: "pending"}, nil
}

func (m *mockOrderRepo) List(_ context.Context, _, _ string, limit, offset int) ([]domain.Order, int, error) {
	m.lastListLimit = limit
	m.lastListOffset = offset
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

// ユーザーが存在しない場合エラーを返し、DeductStockは呼ばれない
func TestCreateOrder_UserNotFound(t *testing.T) {
	t.Parallel()

	uc := NewOrderUsecase(&mockOrderRepo{}, &mockProductClient{
		products: map[string]domain.ProductInfo{
			"p1": {ID: "p1", PriceCents: 1000, StockQuantity: 10},
		},
	}, &mockUserClient{validateErr: errors.New("user not found")})

	_, err := uc.CreateOrder(context.Background(), domain.CreateOrderInput{
		UserID: "unknown",
		Items:  []domain.CreateOrderItem{{ProductID: "p1", Quantity: 1}},
	})

	if err == nil {
		t.Fatal("want error for unknown user, got nil")
	}
}

// 商品が存在しない場合エラーを返す
func TestCreateOrder_ProductNotFound(t *testing.T) {
	t.Parallel()

	uc := NewOrderUsecase(&mockOrderRepo{}, &mockProductClient{products: map[string]domain.ProductInfo{}}, &mockUserClient{})

	_, err := uc.CreateOrder(context.Background(), domain.CreateOrderInput{
		UserID: "u1",
		Items:  []domain.CreateOrderItem{{ProductID: "unknown", Quantity: 1}},
	})

	if err == nil {
		t.Fatal("want error for unknown product, got nil")
	}
}

// 在庫不足の場合 ErrInsufficientStock を返す
func TestCreateOrder_InsufficientStock(t *testing.T) {
	t.Parallel()

	uc := NewOrderUsecase(&mockOrderRepo{}, &mockProductClient{
		products: map[string]domain.ProductInfo{
			"p1": {ID: "p1", Name: "Widget", PriceCents: 1000, StockQuantity: 1},
		},
	}, &mockUserClient{})

	_, err := uc.CreateOrder(context.Background(), domain.CreateOrderInput{
		UserID: "u1",
		Items:  []domain.CreateOrderItem{{ProductID: "p1", Quantity: 5}},
	})

	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("want ErrInsufficientStock, got %v", err)
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

// ---- GetOrder ---------------------------------------------------------------

func TestGetOrder_Success(t *testing.T) {
	t.Parallel()

	repo := &mockOrderRepo{
		orders: map[string]domain.Order{
			"o1": {ID: "o1", Status: "pending"},
		},
	}
	uc := NewOrderUsecase(repo, &mockProductClient{}, &mockUserClient{})

	order, err := uc.GetOrder(context.Background(), "o1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if order.ID != "o1" {
		t.Errorf("ID = %q, want %q", order.ID, "o1")
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockOrderRepo{orders: map[string]domain.Order{}}
	uc := NewOrderUsecase(repo, &mockProductClient{}, &mockUserClient{})

	_, err := uc.GetOrder(context.Background(), "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// ---- ListOrders -------------------------------------------------------------

func TestListOrders_PageNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pageSize, page int
		wantLimit      int
		wantOffset     int
	}{
		{"zero values default to 20/0", 0, 0, 20, 0},
		{"negative values default to 20/0", -1, -1, 20, 0},
		{"explicit page 2 with size 10", 10, 2, 10, 10},
		{"page 3 with size 5", 5, 3, 5, 10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockOrderRepo{}
			uc := NewOrderUsecase(repo, &mockProductClient{}, &mockUserClient{})

			uc.ListOrders(context.Background(), "", "", tt.pageSize, tt.page)

			if repo.lastListLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", repo.lastListLimit, tt.wantLimit)
			}
			if repo.lastListOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", repo.lastListOffset, tt.wantOffset)
			}
		})
	}
}

// ---- UpdateOrderStatus ------------------------------------------------------

func TestUpdateOrderStatus_Success(t *testing.T) {
	t.Parallel()

	uc := NewOrderUsecase(&mockOrderRepo{}, &mockProductClient{}, &mockUserClient{})

	order, err := uc.UpdateOrderStatus(context.Background(), "o1", "confirmed")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if order.Status != "confirmed" {
		t.Errorf("Status = %q, want %q", order.Status, "confirmed")
	}
}

// ---- CancelOrder ------------------------------------------------------------

func TestCancelOrder_Success(t *testing.T) {
	t.Parallel()

	repo := &mockOrderRepo{
		orders: map[string]domain.Order{
			"o1": {
				ID:     "o1",
				Status: "pending",
				Items: []domain.OrderItem{
					{ProductID: "p1", Quantity: 3},
					{ProductID: "p2", Quantity: 1},
				},
			},
		},
	}
	productClient := &mockProductClient{}
	uc := NewOrderUsecase(repo, productClient, &mockUserClient{})

	order, err := uc.CancelOrder(context.Background(), "o1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if order.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", order.Status, "cancelled")
	}
	if len(productClient.restored) != 2 {
		t.Errorf("RestoreStock called %d times, want 2", len(productClient.restored))
	}
}

func TestCancelOrder_OrderNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockOrderRepo{orders: map[string]domain.Order{}}
	uc := NewOrderUsecase(repo, &mockProductClient{}, &mockUserClient{})

	_, err := uc.CancelOrder(context.Background(), "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestCancelOrder_NonPendingOrder(t *testing.T) {
	t.Parallel()

	repo := &mockOrderRepo{
		orders: map[string]domain.Order{
			"o1": {ID: "o1", Status: "confirmed"},
		},
	}
	uc := NewOrderUsecase(repo, &mockProductClient{}, &mockUserClient{})

	_, err := uc.CancelOrder(context.Background(), "o1")

	if !errors.Is(err, domain.ErrInvalidStatus) {
		t.Errorf("want ErrInvalidStatus, got %v", err)
	}
}

func TestCancelOrder_RestoreStockFails(t *testing.T) {
	t.Parallel()

	repo := &mockOrderRepo{
		orders: map[string]domain.Order{
			"o1": {
				ID:     "o1",
				Status: "pending",
				Items:  []domain.OrderItem{{ProductID: "p1", Quantity: 2}},
			},
		},
	}
	productClient := &mockProductClient{
		restoreErrs: map[string]error{"p1": errors.New("restore failed")},
	}
	uc := NewOrderUsecase(repo, productClient, &mockUserClient{})

	_, err := uc.CancelOrder(context.Background(), "o1")

	if err == nil {
		t.Fatal("want error when RestoreStock fails, got nil")
	}
}
