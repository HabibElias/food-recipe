package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/presentation/http/middleware"
)

func TestHasuraWebhookMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name, secret, header string
		status               int
	}{
		{"missing", "private-webhook-key", "", fiber.StatusUnauthorized},
		{"wrong", "private-webhook-key", "different-key", fiber.StatusUnauthorized},
		{"valid", "private-webhook-key", "private-webhook-key", fiber.StatusOK},
		{"local development", "", "", fiber.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/action", middleware.HasuraWebhookMiddleware(tc.secret), func(c *fiber.Ctx) error {
				return c.SendString("accepted")
			})
			req := httptest.NewRequest("POST", "/action", nil)
			req.Header.Set("X-Hasura-Action-Secret", tc.header)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.status {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.status)
			}
		})
	}
}
