package router_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/presentation/http/handlers"
	"go/food-recipe/internal/presentation/http/router"
)

func TestHasuraRoutesRequireWebhookSecret(t *testing.T) {
	app := fiber.New()
	router.SetupRoutes(router.RouterParams{
		App: app, ActionSecret: "private-webhook-key",
		AuthHandler:         handlers.NewAuthHandler(nil),
		PaymentHandler:      handlers.NewPaymentHandler(nil),
		MediaHandler:        handlers.NewMediaHandler(nil),
		NotificationHandler: handlers.NewNotificationHandler(nil),
		HealthHandler:       handlers.NewHealthHandler(),
	})
	for _, path := range []string{"/v1/signup", "/v1/login", "/v1/me", "/v1/payment", "/v1/upload-recipe-images", "/v1/upload-user-image", "/v1/like-notification"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("POST", path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("untrusted request status = %d, want 401", resp.StatusCode)
			}
		})
	}
	for _, path := range []string{"/v1/healthz", "/v1/payment/verify"} {
		t.Run(path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest("GET", path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			want := fiber.StatusOK
			if path == "/v1/payment/verify" {
				want = fiber.StatusBadRequest
			}
			if resp.StatusCode != want {
				t.Fatalf("public route status = %d, want %d", resp.StatusCode, want)
			}
		})
	}
	// A valid Hasura request reaches payload validation instead of the auth gate.
	req := httptest.NewRequest("POST", "/v1/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hasura-Action-Secret", "private-webhook-key")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("trusted invalid payload status = %d, want 400", resp.StatusCode)
	}
}
