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
	if ErrAlreadyExists == nil {
		t.Error("ErrAlreadyExists must not be nil")
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
			wrapped: fmt.Errorf("get product abc: %w", ErrNotFound),
			target:  ErrNotFound,
		},
		{
			name:    "wrapped ErrAlreadyExists is detected",
			wrapped: fmt.Errorf("create product: %w", ErrAlreadyExists),
			target:  ErrAlreadyExists,
		},
		{
			name:    "wrapped ErrInsufficientStock is detected",
			wrapped: fmt.Errorf("update stock: %w", ErrInsufficientStock),
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
	tests := []struct {
		name string
		err  error
		not  error
	}{
		{"ErrNotFound is not ErrAlreadyExists", ErrNotFound, ErrAlreadyExists},
		{"ErrNotFound is not ErrInsufficientStock", ErrNotFound, ErrInsufficientStock},
		{"ErrAlreadyExists is not ErrInsufficientStock", ErrAlreadyExists, ErrInsufficientStock},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errors.Is(tt.err, tt.not) {
				t.Errorf("errors.Is(%v, %v) = true, want false", tt.err, tt.not)
			}
		})
	}
}
