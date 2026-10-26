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
	"go/food-recipe/internal/presentation/response"
)

func TestResponseHelpers(t *testing.T) {
	app := fiber.New()

	app.Get("/ok", func(c *fiber.Ctx) error {
		return response.OK(c, fiber.Map{"key": "value"})
	})
	app.Get("/created", func(c *fiber.Ctx) error {
		return response.Created(c, fiber.Map{"id": 123})
	})
	app.Get("/bad-request", func(c *fiber.Ctx) error {
		return response.BadRequest(c, "bad request err")
	})
	app.Get("/unauthorized", func(c *fiber.Ctx) error {
		return response.Unauthorized(c, "unauthorized err")
	})
	app.Get("/forbidden", func(c *fiber.Ctx) error {
		return response.Forbidden(c, "forbidden err")
	})
	app.Get("/not-found", func(c *fiber.Ctx) error {
		return response.NotFound(c, "not found err")
	})
	app.Get("/conflict", func(c *fiber.Ctx) error {
		return response.Conflict(c, "conflict err")
	})
	app.Get("/internal", func(c *fiber.Ctx) error {
		return response.Internal(c, "internal err")
	})

	tests := []struct {
		route        string
		expectedCode int
		expectedMsg  string
	}{
		{"/ok", fiber.StatusOK, ""},
		{"/created", fiber.StatusCreated, ""},
		{"/bad-request", fiber.StatusBadRequest, "bad request err"},
		{"/unauthorized", fiber.StatusUnauthorized, "unauthorized err"},
		{"/forbidden", fiber.StatusForbidden, "forbidden err"},
		{"/not-found", fiber.StatusNotFound, "not found err"},
		{"/conflict", fiber.StatusConflict, "conflict err"},
		{"/internal", fiber.StatusInternalServerError, "internal err"},
	}

	for _, tt := range tests {
		t.Run(tt.route, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.route, nil)
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
					t.Fatalf("failed to parse error response: %v", err)
				}
				if errResp.Message != tt.expectedMsg {
					t.Errorf("expected message %q, got %q", tt.expectedMsg, errResp.Message)
				}
			}
		})
	}
}

func TestHandleDomainError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "ErrUserNotFound",
			err:          domainerrors.ErrUserNotFound,
			expectedCode: fiber.StatusNotFound,
			expectedMsg:  domainerrors.ErrUserNotFound.Error(),
		},
		{
			name:         "ErrTransactionNotFound wrapped",
			err:          fmt.Errorf("lookup failed: %w", domainerrors.ErrTransactionNotFound),
			expectedCode: fiber.StatusNotFound,
			expectedMsg:  "lookup failed: transaction not found",
		},
		{
			name:         "ErrUserAlreadyExists",
			err:          domainerrors.ErrUserAlreadyExists,
			expectedCode: fiber.StatusConflict,
			expectedMsg:  domainerrors.ErrUserAlreadyExists.Error(),
		},
		{
			name:         "ErrUsernameTaken",
			err:          domainerrors.ErrUsernameTaken,
			expectedCode: fiber.StatusConflict,
			expectedMsg:  domainerrors.ErrUsernameTaken.Error(),
		},
		{
			name:         "ErrInvalidCredentials",
			err:          domainerrors.ErrInvalidCredentials,
			expectedCode: fiber.StatusUnauthorized,
			expectedMsg:  domainerrors.ErrInvalidCredentials.Error(),
		},
		{
			name:         "ErrUnauthorized",
			err:          domainerrors.ErrUnauthorized,
			expectedCode: fiber.StatusUnauthorized,
			expectedMsg:  domainerrors.ErrUnauthorized.Error(),
		},
		{
			name:         "ErrInvalidInput",
			err:          domainerrors.ErrInvalidInput,
			expectedCode: fiber.StatusBadRequest,
			expectedMsg:  domainerrors.ErrInvalidInput.Error(),
		},
		{
			name:         "ErrPaymentFailed",
			err:          domainerrors.ErrPaymentFailed,
			expectedCode: fiber.StatusBadRequest,
			expectedMsg:  domainerrors.ErrPaymentFailed.Error(),
		},
		{
			name:         "ErrServiceUnavailable",
			err:          domainerrors.ErrServiceUnavailable,
			expectedCode: fiber.StatusServiceUnavailable,
			expectedMsg:  domainerrors.ErrServiceUnavailable.Error(),
		},
		{
			name:         "Default unknown error",
			err:          errors.New("something totally unknown"),
			expectedCode: fiber.StatusInternalServerError,
			expectedMsg:  "something totally unknown",
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
