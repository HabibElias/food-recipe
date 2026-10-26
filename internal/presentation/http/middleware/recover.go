package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/presentation/response"
)

// RecoverMiddleware recovers from panics and returns a 500 error response.
func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERED] %v\n%s", r, debug.Stack())
				_ = response.Internal(c, "internal server error")
			}
		}()
		return c.Next()
	}
}
