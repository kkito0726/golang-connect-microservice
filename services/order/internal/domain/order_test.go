package domain

import (
	"errors"
	"testing"
)

// ---- NewOrder ---------------------------------------------------------------

func TestNewOrder_Success(t *testing.T) {
	t.Parallel()

	products := map[string]ProductInfo{
		"p1": {ID: "p1", Name: "Widget", PriceCents: 1000, StockQuantity: 10},
		"p2": {ID: "p2", Name: "Gadget", PriceCents: 2000, StockQuantity: 5},
	}
	items := []CreateOrderItem{
		{ProductID: "p1", Quantity: 2},
		{ProductID: "p2", Quantity: 1},
	}

	order, err := NewOrder("u1", items, products)

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if order.UserID != "u1" {
		t.Errorf("UserID = %q, want %q", order.UserID, "u1")
	}
	if order.Status != "pending" {
		t.Errorf("Status = %q, want %q", order.Status, "pending")
	}
	if order.TotalCents != 4000 {
		t.Errorf("TotalCents = %d, want 4000", order.TotalCents)
	}
	if len(order.Items) != 2 {
		t.Errorf("len(Items) = %d, want 2", len(order.Items))
	}
}

func TestNewOrder_InsufficientStock(t *testing.T) {
	t.Parallel()

	products := map[string]ProductInfo{
		"p1": {ID: "p1", Name: "Widget", PriceCents: 1000, StockQuantity: 1},
	}
	items := []CreateOrderItem{
		{ProductID: "p1", Quantity: 5},
	}

	_, err := NewOrder("u1", items, products)

	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("want ErrInsufficientStock, got %v", err)
	}
}

func TestNewOrder_ProductNotFound(t *testing.T) {
	t.Parallel()

	_, err := NewOrder("u1", []CreateOrderItem{{ProductID: "unknown", Quantity: 1}}, map[string]ProductInfo{})

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestNewOrder_TotalCentsCalculation(t *testing.T) {
	t.Parallel()

	products := map[string]ProductInfo{
		"p1": {ID: "p1", PriceCents: 300, StockQuantity: 10},
		"p2": {ID: "p2", PriceCents: 700, StockQuantity: 10},
	}
	items := []CreateOrderItem{
		{ProductID: "p1", Quantity: 3},
		{ProductID: "p2", Quantity: 2},
	}

	order, err := NewOrder("u1", items, products)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	// 300*3 + 700*2 = 900 + 1400 = 2300
	if order.TotalCents != 2300 {
		t.Errorf("TotalCents = %d, want 2300", order.TotalCents)
	}
}

// ---- Cancel -----------------------------------------------------------------

func TestCancel_PendingOrder(t *testing.T) {
	t.Parallel()

	order := Order{ID: "o1", Status: "pending"}

	cancelled, err := order.Cancel()

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", cancelled.Status, "cancelled")
	}
	if order.Status != "pending" {
		t.Error("original order must not be mutated")
	}
}

func TestCancel_NonPendingOrder(t *testing.T) {
	t.Parallel()

	statuses := []string{"cancelled", "completed", "shipped", ""}
	for _, s := range statuses {
		s := s
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			_, err := Order{ID: "o1", Status: s}.Cancel()
			if !errors.Is(err, ErrInvalidStatus) {
				t.Errorf("status %q: want ErrInvalidStatus, got %v", s, err)
			}
		})
	}
}
