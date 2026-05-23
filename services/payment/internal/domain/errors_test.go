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
			wrapped: fmt.Errorf("get payment abc: %w", ErrNotFound),
			target:  ErrNotFound,
		},
		{
			name:    "wrapped ErrInvalidStatus is detected",
			wrapped: fmt.Errorf("refund payment: %w", ErrInvalidStatus),
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
	if errors.Is(ErrNotFound, ErrInvalidStatus) {
		t.Error("ErrNotFound must not match ErrInvalidStatus")
	}
	if errors.Is(ErrInvalidStatus, ErrNotFound) {
		t.Error("ErrInvalidStatus must not match ErrNotFound")
	}
}

func TestSentinelErrors_DirectMatchIsNotRequired(t *testing.T) {
	other := errors.New("not found")
	if errors.Is(other, ErrNotFound) {
		t.Error("a different 'not found' error must not match domain.ErrNotFound")
	}
}
