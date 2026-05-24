package domain

import (
	"errors"
	"testing"
)

func TestApplyStockDelta_AddStock(t *testing.T) {
	t.Parallel()

	p := Product{StockQuantity: 10}
	updated, err := p.ApplyStockDelta(5)

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if updated.StockQuantity != 15 {
		t.Errorf("StockQuantity = %d, want 15", updated.StockQuantity)
	}
	if p.StockQuantity != 10 {
		t.Error("original product must not be mutated")
	}
}

func TestApplyStockDelta_DeductStock(t *testing.T) {
	t.Parallel()

	p := Product{StockQuantity: 10}
	updated, err := p.ApplyStockDelta(-3)

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if updated.StockQuantity != 7 {
		t.Errorf("StockQuantity = %d, want 7", updated.StockQuantity)
	}
}

func TestApplyStockDelta_ExactDepletion(t *testing.T) {
	t.Parallel()

	p := Product{StockQuantity: 5}
	updated, err := p.ApplyStockDelta(-5)

	if err != nil {
		t.Fatalf("want no error when stock reaches exactly zero, got %v", err)
	}
	if updated.StockQuantity != 0 {
		t.Errorf("StockQuantity = %d, want 0", updated.StockQuantity)
	}
}

func TestApplyStockDelta_InsufficientStock(t *testing.T) {
	t.Parallel()

	p := Product{StockQuantity: 3}
	_, err := p.ApplyStockDelta(-5)

	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("want ErrInsufficientStock, got %v", err)
	}
}

func TestApplyStockDelta_ZeroDelta(t *testing.T) {
	t.Parallel()

	p := Product{StockQuantity: 10}
	updated, err := p.ApplyStockDelta(0)

	if err != nil {
		t.Fatalf("want no error for zero delta, got %v", err)
	}
	if updated.StockQuantity != 10 {
		t.Errorf("StockQuantity = %d, want 10", updated.StockQuantity)
	}
}
