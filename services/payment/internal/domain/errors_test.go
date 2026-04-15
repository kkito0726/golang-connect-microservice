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
}

func TestSentinelErrors_IsDetectsWrapped(t *testing.T) {
	wrapped := fmt.Errorf("get payment abc: %w", ErrNotFound)
	if !errors.Is(wrapped, ErrNotFound) {
		t.Errorf("errors.Is(%v, ErrNotFound) = false, want true", wrapped)
	}
}

func TestSentinelErrors_DirectMatchIsNotRequired(t *testing.T) {
	other := errors.New("not found")
	if errors.Is(other, ErrNotFound) {
		t.Error("a different 'not found' error must not match domain.ErrNotFound")
	}
}
