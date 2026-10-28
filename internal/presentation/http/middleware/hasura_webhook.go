package middleware

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/presentation/response"
)

// HasuraWebhookMiddleware authenticates action and event requests from Hasura.
// An empty secret preserves the existing local-development workflow.
func HasuraWebhookMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if secret != "" && subtle.ConstantTimeCompare([]byte(c.Get("X-Hasura-Action-Secret")), []byte(secret)) != 1 {
			return response.Unauthorized(c, "unauthorized webhook request")
		}
		return c.Next()
	}
}
