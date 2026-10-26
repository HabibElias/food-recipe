package errors_test

import (
	"testing"

	"go/food-recipe/internal/domain/errors"
)

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		err     error
		message string
	}{
		{errors.ErrUserNotFound, "user not found"},
		{errors.ErrUserAlreadyExists, "email already registered"},
		{errors.ErrUsernameTaken, "username already taken"},
		{errors.ErrInvalidCredentials, "invalid credentials"},
		{errors.ErrUnauthorized, "unauthorized"},
		{errors.ErrInvalidInput, "invalid input"},
		{errors.ErrTransactionNotFound, "transaction not found"},
		{errors.ErrPaymentFailed, "payment failed"},
		{errors.ErrServiceUnavailable, "service temporarily unavailable"},
	}

	for _, tt := range tests {
		if tt.err == nil {
			t.Fatalf("expected error to not be nil")
		}
		if tt.err.Error() != tt.message {
			t.Errorf("expected %q, got %q", tt.message, tt.err.Error())
		}
	}
}
