package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ken/connect-microservice/services/product/internal/domain"
)

func TestSentinelErrors_NotNil(t *testing.T) {
	if domain.ErrNotFound == nil {
		t.Error("ErrNotFound must not be nil")
	}
	if domain.ErrAlreadyExists == nil {
		t.Error("ErrAlreadyExists must not be nil")
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
			wrapped: fmt.Errorf("get product abc: %w", domain.ErrNotFound),
			target:  domain.ErrNotFound,
		},
		{
			name:    "wrapped ErrAlreadyExists is detected",
			wrapped: fmt.Errorf("create product: %w", domain.ErrAlreadyExists),
			target:  domain.ErrAlreadyExists,
		},
		{
			name:    "wrapped ErrInsufficientStock is detected",
			wrapped: fmt.Errorf("update stock: %w", domain.ErrInsufficientStock),
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
	tests := []struct {
		name string
		err  error
		not  error
	}{
		{"ErrNotFound is not ErrAlreadyExists", domain.ErrNotFound, domain.ErrAlreadyExists},
		{"ErrNotFound is not ErrInsufficientStock", domain.ErrNotFound, domain.ErrInsufficientStock},
		{"ErrAlreadyExists is not ErrInsufficientStock", domain.ErrAlreadyExists, domain.ErrInsufficientStock},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errors.Is(tt.err, tt.not) {
				t.Errorf("errors.Is(%v, %v) = true, want false", tt.err, tt.not)
			}
		})
	}
}
