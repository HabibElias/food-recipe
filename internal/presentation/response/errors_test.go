package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/pkg/validator"
	"go/food-recipe/internal/presentation/response"
)

func TestHandleDomainError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "ErrUserNotFound raw",
			err:          domainerrors.ErrUserNotFound,
			expectedCode: fiber.StatusNotFound,
			expectedMsg:  domainerrors.ErrUserNotFound.Error(),
		},
		{
			name:         "ErrUserNotFound wrapped",
			err:          fmt.Errorf("user lookup failed: %w", domainerrors.ErrUserNotFound),
			expectedCode: fiber.StatusNotFound,
			expectedMsg:  fmt.Sprintf("user lookup failed: %s", domainerrors.ErrUserNotFound.Error()),
		},
		{
			name:         "ErrTransactionNotFound raw",
			err:          domainerrors.ErrTransactionNotFound,
			expectedCode: fiber.StatusNotFound,
			expectedMsg:  domainerrors.ErrTransactionNotFound.Error(),
		},
		{
			name:         "ErrTransactionNotFound wrapped",
			err:          fmt.Errorf("tx lookup failed: %w", domainerrors.ErrTransactionNotFound),
			expectedCode: fiber.StatusNotFound,
			expectedMsg:  fmt.Sprintf("tx lookup failed: %s", domainerrors.ErrTransactionNotFound.Error()),
		},
		{
			name:         "ErrUserAlreadyExists raw",
			err:          domainerrors.ErrUserAlreadyExists,
			expectedCode: fiber.StatusConflict,
			expectedMsg:  domainerrors.ErrUserAlreadyExists.Error(),
		},
		{
			name:         "ErrUserAlreadyExists wrapped",
			err:          fmt.Errorf("conflict: %w", domainerrors.ErrUserAlreadyExists),
			expectedCode: fiber.StatusConflict,
			expectedMsg:  fmt.Sprintf("conflict: %s", domainerrors.ErrUserAlreadyExists.Error()),
		},
		{
			name:         "ErrUsernameTaken raw",
			err:          domainerrors.ErrUsernameTaken,
			expectedCode: fiber.StatusConflict,
			expectedMsg:  domainerrors.ErrUsernameTaken.Error(),
		},
		{
			name:         "ErrUsernameTaken wrapped",
			err:          fmt.Errorf("conflict: %w", domainerrors.ErrUsernameTaken),
			expectedCode: fiber.StatusConflict,
			expectedMsg:  fmt.Sprintf("conflict: %s", domainerrors.ErrUsernameTaken.Error()),
		},
		{
			name:         "ErrInvalidCredentials raw",
			err:          domainerrors.ErrInvalidCredentials,
			expectedCode: fiber.StatusUnauthorized,
			expectedMsg:  domainerrors.ErrInvalidCredentials.Error(),
		},
		{
			name:         "ErrInvalidCredentials wrapped",
			err:          fmt.Errorf("auth error: %w", domainerrors.ErrInvalidCredentials),
			expectedCode: fiber.StatusUnauthorized,
			expectedMsg:  fmt.Sprintf("auth error: %s", domainerrors.ErrInvalidCredentials.Error()),
		},
		{
			name:         "ErrUnauthorized raw",
			err:          domainerrors.ErrUnauthorized,
			expectedCode: fiber.StatusUnauthorized,
			expectedMsg:  domainerrors.ErrUnauthorized.Error(),
		},
		{
			name:         "ErrUnauthorized wrapped",
			err:          fmt.Errorf("guard error: %w", domainerrors.ErrUnauthorized),
			expectedCode: fiber.StatusUnauthorized,
			expectedMsg:  fmt.Sprintf("guard error: %s", domainerrors.ErrUnauthorized.Error()),
		},
		{
			name:         "ErrInvalidInput raw",
			err:          domainerrors.ErrInvalidInput,
			expectedCode: fiber.StatusBadRequest,
			expectedMsg:  domainerrors.ErrInvalidInput.Error(),
		},
		{
			name:         "ErrInvalidInput wrapped",
			err:          fmt.Errorf("bad payload: %w", domainerrors.ErrInvalidInput),
			expectedCode: fiber.StatusBadRequest,
			expectedMsg:  fmt.Sprintf("bad payload: %s", domainerrors.ErrInvalidInput.Error()),
		},
		{
			name:         "ErrPaymentFailed raw",
			err:          domainerrors.ErrPaymentFailed,
			expectedCode: fiber.StatusBadRequest,
			expectedMsg:  domainerrors.ErrPaymentFailed.Error(),
		},
		{
			name:         "ErrPaymentFailed wrapped",
			err:          fmt.Errorf("gateway error: %w", domainerrors.ErrPaymentFailed),
			expectedCode: fiber.StatusBadRequest,
			expectedMsg:  fmt.Sprintf("gateway error: %s", domainerrors.ErrPaymentFailed.Error()),
		},
		{
			name:         "ErrServiceUnavailable raw",
			err:          domainerrors.ErrServiceUnavailable,
			expectedCode: fiber.StatusServiceUnavailable,
			expectedMsg:  domainerrors.ErrServiceUnavailable.Error(),
		},
		{
			name:         "ErrServiceUnavailable wrapped",
			err:          fmt.Errorf("upstream error: %w", domainerrors.ErrServiceUnavailable),
			expectedCode: fiber.StatusServiceUnavailable,
			expectedMsg:  fmt.Sprintf("upstream error: %s", domainerrors.ErrServiceUnavailable.Error()),
		},
		{
			name:         "Default unknown error",
			err:          errors.New("something totally unknown"),
			expectedCode: fiber.StatusInternalServerError,
			expectedMsg:  "internal server error",
		},
		{
			name:         "Nil error",
			err:          nil,
			expectedCode: fiber.StatusOK,
			expectedMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				return response.HandleDomainError(c, tt.err)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			if tt.expectedMsg != "" {
				body, _ := io.ReadAll(resp.Body)
				var errResp response.ErrorResponse
				if err := json.Unmarshal(body, &errResp); err != nil {
					t.Fatalf("failed to parse json: %v", err)
				}
				if errResp.Message != tt.expectedMsg {
					t.Errorf("expected message %q, got %q", tt.expectedMsg, errResp.Message)
				}
			}
		})
	}
}

func TestHandleDomainError_ValidationError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		valErr := &validator.ValidationError{
			Errors: []validator.FieldError{
				{Field: "email", Error: "email is required"},
				{Field: "password", Error: "password must be at least 6 characters"},
			},
		}
		return response.HandleDomainError(c, valErr)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Message string                 `json:"message"`
		Errors  []validator.FieldError `json:"errors"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if out.Message != "email is required, password must be at least 6 characters" {
		t.Errorf("expected message %q, got %q", "email is required, password must be at least 6 characters", out.Message)
	}
	if len(out.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(out.Errors))
	}
	if out.Errors[0].Field != "email" || out.Errors[0].Error != "email is required" {
		t.Errorf("unexpected first error: %+v", out.Errors[0])
	}
	if out.Errors[1].Field != "password" || out.Errors[1].Error != "password must be at least 6 characters" {
		t.Errorf("unexpected second error: %+v", out.Errors[1])
	}
}
