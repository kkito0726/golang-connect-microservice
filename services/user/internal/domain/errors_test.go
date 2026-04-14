package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ken/connect-microservice/services/user/internal/domain"
)

func TestSentinelErrors_NotNil(t *testing.T) {
	if domain.ErrNotFound == nil {
		t.Error("ErrNotFound must not be nil")
	}
	if domain.ErrAlreadyExists == nil {
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
			wrapped: fmt.Errorf("get user abc: %w", domain.ErrNotFound),
			target:  domain.ErrNotFound,
		},
		{
			name:    "wrapped ErrAlreadyExists is detected",
			wrapped: fmt.Errorf("create user: %w", domain.ErrAlreadyExists),
			target:  domain.ErrAlreadyExists,
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
	if errors.Is(domain.ErrNotFound, domain.ErrAlreadyExists) {
		t.Error("ErrNotFound must not match ErrAlreadyExists")
	}
	if errors.Is(domain.ErrAlreadyExists, domain.ErrNotFound) {
		t.Error("ErrAlreadyExists must not match ErrNotFound")
	}
}
