package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ken/connect-microservice/services/payment/internal/domain"
)

func TestSentinelErrors_NotNil(t *testing.T) {
	if domain.ErrNotFound == nil {
		t.Error("ErrNotFound must not be nil")
	}
}

func TestSentinelErrors_IsDetectsWrapped(t *testing.T) {
	wrapped := fmt.Errorf("get payment abc: %w", domain.ErrNotFound)
	if !errors.Is(wrapped, domain.ErrNotFound) {
		t.Errorf("errors.Is(%v, domain.ErrNotFound) = false, want true", wrapped)
	}
}

func TestSentinelErrors_DirectMatchIsNotRequired(t *testing.T) {
	other := errors.New("not found")
	if errors.Is(other, domain.ErrNotFound) {
		t.Error("a different 'not found' error must not match domain.ErrNotFound")
	}
}
