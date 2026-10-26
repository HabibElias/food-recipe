package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/presentation/response"
)

// HealthHandler provides health check and diagnostic endpoints.
type HealthHandler struct{}

// NewHealthHandler constructs a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Healthz returns 200 OK with server health status.
func (h *HealthHandler) Healthz(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{"status": "ok"})
}

// Err returns 500 Internal Server Error for testing error responses.
func (h *HealthHandler) Err(c *fiber.Ctx) error {
	return response.Internal(c, "Internal Server Error")
}
