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
}

func TestSentinelErrors_IsDetectsWrapped(t *testing.T) {
	tests := []struct {
		name    string
		wrapped error
		target  error
	}{
		{
			name:    "wrapped ErrNotFound is detected",
			wrapped: fmt.Errorf("get user abc: %w", ErrNotFound),
			target:  ErrNotFound,
		},
		{
			name:    "wrapped ErrAlreadyExists is detected",
			wrapped: fmt.Errorf("create user: %w", ErrAlreadyExists),
			target:  ErrAlreadyExists,
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
	if errors.Is(ErrNotFound, ErrAlreadyExists) {
		t.Error("ErrNotFound must not match ErrAlreadyExists")
	}
	if errors.Is(ErrAlreadyExists, ErrNotFound) {
		t.Error("ErrAlreadyExists must not match ErrNotFound")
	}
}
