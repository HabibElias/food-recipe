package response_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/infrastructure/validator"
	"go/food-recipe/internal/presentation/response"
)

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
