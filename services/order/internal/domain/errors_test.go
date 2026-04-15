package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_NotNil(t *testing.T) {
	if ErrNotFound == nil {
		t.Error("ErrNotFound must not be nil")
	}
	if ErrInsufficientStock == nil {
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
			wrapped: fmt.Errorf("get order abc: %w", ErrNotFound),
			target:  ErrNotFound,
		},
		{
			name:    "wrapped ErrInsufficientStock is detected",
			wrapped: fmt.Errorf("deduct stock: %w", ErrInsufficientStock),
			target:  ErrInsufficientStock,
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
	if errors.Is(ErrNotFound, ErrInsufficientStock) {
		t.Error("ErrNotFound must not match ErrInsufficientStock")
	}
	if errors.Is(ErrInsufficientStock, ErrNotFound) {
		t.Error("ErrInsufficientStock must not match ErrNotFound")
	}
}
