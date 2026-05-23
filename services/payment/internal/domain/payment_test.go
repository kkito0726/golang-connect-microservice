package domain

import (
	"errors"
	"testing"
)

// ---- NewPayment -------------------------------------------------------------

func TestNewPayment_SetsFieldsCorrectly(t *testing.T) {
	t.Parallel()

	p := NewPayment("order-1", "user-1", "credit_card", 5000)

	if p.OrderID != "order-1" {
		t.Errorf("OrderID = %q, want %q", p.OrderID, "order-1")
	}
	if p.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", p.UserID, "user-1")
	}
	if p.Method != "credit_card" {
		t.Errorf("Method = %q, want %q", p.Method, "credit_card")
	}
	if p.AmountCents != 5000 {
		t.Errorf("AmountCents = %d, want 5000", p.AmountCents)
	}
	if p.Status != "completed" {
		t.Errorf("Status = %q, want %q", p.Status, "completed")
	}
}

// ---- Refund -----------------------------------------------------------------

func TestRefund_CompletedPayment(t *testing.T) {
	t.Parallel()

	p := Payment{ID: "pay-1", Status: "completed"}

	refunded, err := p.Refund()

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if refunded.Status != "refunded" {
		t.Errorf("Status = %q, want %q", refunded.Status, "refunded")
	}
	if p.Status != "completed" {
		t.Error("original payment must not be mutated")
	}
}

func TestRefund_NonCompletedPayment(t *testing.T) {
	t.Parallel()

	statuses := []string{"refunded", "pending", "failed", ""}
	for _, s := range statuses {
		s := s
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			_, err := Payment{ID: "pay-1", Status: s}.Refund()
			if !errors.Is(err, ErrInvalidStatus) {
				t.Errorf("status %q: want ErrInvalidStatus, got %v", s, err)
			}
		})
	}
}
