package middleware_test

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go/food-recipe/internal/infrastructure/security"
	"go/food-recipe/internal/presentation/http/middleware"
)

func TestLoggerMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.LoggerMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRecoverMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RecoverMiddleware())
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("something went terribly wrong")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected 500 status on panic, got %d", resp.StatusCode)
	}
}

func TestCORSMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.CORSMiddleware("https://example.com"))
	app.Get("/cors", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("OPTIONS", "/cors", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "https://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin to be https://example.com, got %q", origin)
	}
}

func TestAuthMiddleware(t *testing.T) {
	secret := "super-secure-secret-key"
	userID := uuid.New().String()
	email := "test@example.com"

	app := fiber.New()
	app.Use(middleware.AuthMiddleware(secret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		uid := c.Locals("user_id")
		return c.JSON(fiber.Map{"user_id": uid})
	})

	t.Run("MissingHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Token abcxyz")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("ValidToken", func(t *testing.T) {
		token, err := security.GenerateHasuraJWT(userID, email, secret)
		if err != nil {
			t.Fatalf("failed to generate JWT: %v", err)
		}

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if !bytes.Contains(body, []byte(userID)) {
			t.Errorf("expected body to contain user ID %s, got %s", userID, string(body))
		}
	})
}
