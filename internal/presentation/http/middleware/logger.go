package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LoggerMiddleware logs incoming requests with method, path, status, and duration.
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Printf("[%s] %s - %d (%v)", c.Method(), c.Path(), c.Response().StatusCode(), time.Since(start))
		return err
	}
}
