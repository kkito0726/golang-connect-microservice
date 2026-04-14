package handler

import (
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"

	"github.com/ken/connect-microservice/services/product/internal/domain"
)

func TestToConnectError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantCode     connect.Code
		wantMsgExact string
	}{
		{
			name:     "ErrNotFound maps to CodeNotFound",
			err:      domain.ErrNotFound,
			wantCode: connect.CodeNotFound,
		},
		{
			name:     "wrapped ErrNotFound maps to CodeNotFound",
			err:      fmt.Errorf("get product abc: %w", domain.ErrNotFound),
			wantCode: connect.CodeNotFound,
		},
		{
			name:     "ErrAlreadyExists maps to CodeAlreadyExists",
			err:      domain.ErrAlreadyExists,
			wantCode: connect.CodeAlreadyExists,
		},
		{
			name:     "wrapped ErrAlreadyExists maps to CodeAlreadyExists",
			err:      fmt.Errorf("create product: %w", domain.ErrAlreadyExists),
			wantCode: connect.CodeAlreadyExists,
		},
		{
			name:     "ErrInsufficientStock maps to CodeFailedPrecondition",
			err:      domain.ErrInsufficientStock,
			wantCode: connect.CodeFailedPrecondition,
		},
		{
			name:     "wrapped ErrInsufficientStock maps to CodeFailedPrecondition",
			err:      fmt.Errorf("update stock: %w", domain.ErrInsufficientStock),
			wantCode: connect.CodeFailedPrecondition,
		},
		{
			name:         "unknown error maps to CodeInternal with generic message",
			err:          errors.New("db connection refused"),
			wantCode:     connect.CodeInternal,
			wantMsgExact: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toConnectError(tt.err)

			var connectErr *connect.Error
			if !errors.As(got, &connectErr) {
				t.Fatalf("toConnectError(%v) did not return *connect.Error", tt.err)
			}

			if connectErr.Code() != tt.wantCode {
				t.Errorf("code = %v, want %v", connectErr.Code(), tt.wantCode)
			}

			if tt.wantMsgExact != "" && connectErr.Message() != tt.wantMsgExact {
				t.Errorf("message = %q, want %q", connectErr.Message(), tt.wantMsgExact)
			}
		})
	}
}

func TestToConnectError_InternalDoesNotLeakDetails(t *testing.T) {
	sensitiveErr := errors.New("password=secret123 connection failed")
	got := toConnectError(sensitiveErr)

	var connectErr *connect.Error
	if !errors.As(got, &connectErr) {
		t.Fatal("expected *connect.Error")
	}

	if connectErr.Message() == sensitiveErr.Error() {
		t.Error("internal error must not expose the original error message to callers")
	}
}
