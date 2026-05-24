package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ken/connect-microservice/services/product/internal/domain"
)

// ---- モック ----------------------------------------------------------------

type mockProductRepo struct {
	products       map[string]domain.Product
	movements      map[string][]domain.StockMovement
	createFn       func(ctx context.Context, p domain.Product) (domain.Product, error)
	updateStockFn  func(ctx context.Context, productID string, delta int32, reason, referenceID string) (domain.Product, domain.StockMovement, error)
	lastListLimit  int
	lastListOffset int
}

func (m *mockProductRepo) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	p.ID = "prod-1"
	return p, nil
}

func (m *mockProductRepo) GetByID(_ context.Context, id string) (domain.Product, error) {
	if m.products == nil {
		return domain.Product{}, domain.ErrNotFound
	}
	p, ok := m.products[id]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockProductRepo) List(_ context.Context, limit, offset int, _ string) ([]domain.Product, int, error) {
	m.lastListLimit = limit
	m.lastListOffset = offset
	return nil, 0, nil
}

func (m *mockProductRepo) Update(_ context.Context, id, name, description, category string, priceCents int64) (domain.Product, error) {
	return domain.Product{ID: id, Name: name, Description: description, Category: category, PriceCents: priceCents}, nil
}

func (m *mockProductRepo) SoftDelete(_ context.Context, id string) error {
	if m.products == nil {
		return domain.ErrNotFound
	}
	if _, ok := m.products[id]; !ok {
		return domain.ErrNotFound
	}
	return nil
}

func (m *mockProductRepo) UpdateStock(ctx context.Context, productID string, delta int32, reason, referenceID string) (domain.Product, domain.StockMovement, error) {
	if m.updateStockFn != nil {
		return m.updateStockFn(ctx, productID, delta, reason, referenceID)
	}
	return domain.Product{ID: productID}, domain.StockMovement{ProductID: productID, Delta: delta}, nil
}

func (m *mockProductRepo) GetStockMovements(_ context.Context, productID string, _ int) ([]domain.StockMovement, error) {
	if m.movements == nil {
		return nil, errors.New("movements not found")
	}
	mvs, ok := m.movements[productID]
	if !ok {
		return []domain.StockMovement{}, nil
	}
	return mvs, nil
}

// ---- CreateProduct ----------------------------------------------------------

func TestCreateProduct_Success(t *testing.T) {
	t.Parallel()

	uc := NewProductUsecase(&mockProductRepo{})

	p, err := uc.CreateProduct(context.Background(), domain.Product{Name: "Widget", SKU: "W-001"})

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if p.ID == "" {
		t.Error("want non-empty ID after create")
	}
}

// ---- GetProduct -------------------------------------------------------------

func TestGetProduct_Success(t *testing.T) {
	t.Parallel()

	repo := &mockProductRepo{
		products: map[string]domain.Product{
			"p1": {ID: "p1", Name: "Widget"},
		},
	}
	uc := NewProductUsecase(repo)

	p, err := uc.GetProduct(context.Background(), "p1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if p.Name != "Widget" {
		t.Errorf("Name = %q, want %q", p.Name, "Widget")
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	t.Parallel()

	uc := NewProductUsecase(&mockProductRepo{products: map[string]domain.Product{}})

	_, err := uc.GetProduct(context.Background(), "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// ---- ListProducts -----------------------------------------------------------

func TestListProducts_PageNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pageSize, page int
		wantLimit      int
		wantOffset     int
	}{
		{"zero values default to 20/0", 0, 0, 20, 0},
		{"negative values default to 20/0", -1, -1, 20, 0},
		{"explicit page 2 with size 5", 5, 2, 5, 5},
		{"page 3 with size 10", 10, 3, 10, 20},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockProductRepo{}
			uc := NewProductUsecase(repo)

			uc.ListProducts(context.Background(), tt.pageSize, tt.page, "")

			if repo.lastListLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", repo.lastListLimit, tt.wantLimit)
			}
			if repo.lastListOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", repo.lastListOffset, tt.wantOffset)
			}
		})
	}
}

// ---- UpdateProduct ----------------------------------------------------------

func TestUpdateProduct_Success(t *testing.T) {
	t.Parallel()

	uc := NewProductUsecase(&mockProductRepo{})

	p, err := uc.UpdateProduct(context.Background(), "p1", "NewName", "desc", "cat", 2000)

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if p.Name != "NewName" {
		t.Errorf("Name = %q, want %q", p.Name, "NewName")
	}
	if p.PriceCents != 2000 {
		t.Errorf("PriceCents = %d, want 2000", p.PriceCents)
	}
}

// ---- DeleteProduct ----------------------------------------------------------

func TestDeleteProduct_Success(t *testing.T) {
	t.Parallel()

	repo := &mockProductRepo{
		products: map[string]domain.Product{"p1": {ID: "p1"}},
	}
	uc := NewProductUsecase(repo)

	err := uc.DeleteProduct(context.Background(), "p1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func TestDeleteProduct_NotFound(t *testing.T) {
	t.Parallel()

	uc := NewProductUsecase(&mockProductRepo{products: map[string]domain.Product{}})

	err := uc.DeleteProduct(context.Background(), "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// ---- UpdateStock ------------------------------------------------------------

func TestUpdateStock_Success(t *testing.T) {
	t.Parallel()

	uc := NewProductUsecase(&mockProductRepo{})

	p, mv, err := uc.UpdateStock(context.Background(), "p1", -3, "sale", "ref-1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if p.ID != "p1" {
		t.Errorf("Product.ID = %q, want %q", p.ID, "p1")
	}
	if mv.Delta != -3 {
		t.Errorf("StockMovement.Delta = %d, want -3", mv.Delta)
	}
}

// ---- GetStockLevel ----------------------------------------------------------

func TestGetStockLevel_Success(t *testing.T) {
	t.Parallel()

	repo := &mockProductRepo{
		products: map[string]domain.Product{
			"p1": {ID: "p1", Name: "Widget", StockQuantity: 42},
		},
		movements: map[string][]domain.StockMovement{
			"p1": {
				{ID: "m1", ProductID: "p1", Delta: -5, Reason: "sale"},
				{ID: "m2", ProductID: "p1", Delta: 10, Reason: "restock"},
			},
		},
	}
	uc := NewProductUsecase(repo)

	p, mvs, err := uc.GetStockLevel(context.Background(), "p1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if p.StockQuantity != 42 {
		t.Errorf("StockQuantity = %d, want 42", p.StockQuantity)
	}
	if len(mvs) != 2 {
		t.Errorf("len(movements) = %d, want 2", len(mvs))
	}
}

func TestGetStockLevel_ProductNotFound(t *testing.T) {
	t.Parallel()

	uc := NewProductUsecase(&mockProductRepo{products: map[string]domain.Product{}})

	_, _, err := uc.GetStockLevel(context.Background(), "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestGetStockLevel_MovementsError(t *testing.T) {
	t.Parallel()

	// products に p1 あり、movements は nil（GetStockMovements がエラーを返す）
	repo := &mockProductRepo{
		products: map[string]domain.Product{
			"p1": {ID: "p1"},
		},
	}
	uc := NewProductUsecase(repo)

	_, _, err := uc.GetStockLevel(context.Background(), "p1")

	if err == nil {
		t.Fatal("want error when GetStockMovements fails, got nil")
	}
}
