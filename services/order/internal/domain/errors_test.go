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
	if ErrInvalidStatus == nil {
		t.Error("ErrInvalidStatus must not be nil")
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
		{
			name:    "wrapped ErrInvalidStatus is detected",
			wrapped: fmt.Errorf("cancel order: %w", ErrInvalidStatus),
			target:  ErrInvalidStatus,
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
	pairs := [][2]error{
		{ErrNotFound, ErrInsufficientStock},
		{ErrNotFound, ErrInvalidStatus},
		{ErrInsufficientStock, ErrInvalidStatus},
	}
	for _, p := range pairs {
		if errors.Is(p[0], p[1]) {
			t.Errorf("%v must not match %v", p[0], p[1])
		}
		if errors.Is(p[1], p[0]) {
			t.Errorf("%v must not match %v", p[1], p[0])
		}
	}
}
