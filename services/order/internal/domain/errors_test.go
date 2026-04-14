package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ken/connect-microservice/services/order/internal/domain"
)

func TestSentinelErrors_NotNil(t *testing.T) {
	if domain.ErrNotFound == nil {
		t.Error("ErrNotFound must not be nil")
	}
	if domain.ErrInsufficientStock == nil {
		t.Error("ErrInsufficientStock must not be nil")
	}
}

func TestSentinelErrors_IsDetectsWrapped(t *testing.T) {
	tests := []struct {
		name    string
		wrapped error
		target  error
	}{
		{
			name:    "wrapped ErrNotFound is detected",
			wrapped: fmt.Errorf("get order abc: %w", domain.ErrNotFound),
			target:  domain.ErrNotFound,
		},
		{
			name:    "wrapped ErrInsufficientStock is detected",
			wrapped: fmt.Errorf("deduct stock: %w", domain.ErrInsufficientStock),
			target:  domain.ErrInsufficientStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.wrapped, tt.target) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.wrapped, tt.target)
			}
		})
	}
}

func TestSentinelErrors_IsDistinct(t *testing.T) {
	if errors.Is(domain.ErrNotFound, domain.ErrInsufficientStock) {
		t.Error("ErrNotFound must not match ErrInsufficientStock")
	}
	if errors.Is(domain.ErrInsufficientStock, domain.ErrNotFound) {
		t.Error("ErrInsufficientStock must not match ErrNotFound")
	}
}
