package response

import "github.com/gofiber/fiber/v2"

// SuccessResponse represents a standard successful JSON response wrapper.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents a JSON error response matching Hasura Action format.
type ErrorResponse struct {
	Message string `json:"message"`
}

// JSON sends a JSON response with status code.
func JSON(c *fiber.Ctx, status int, payload interface{}) error {
	return c.Status(status).JSON(payload)
}

// OK sends 200 OK with payload.
func OK(c *fiber.Ctx, payload interface{}) error {
	return JSON(c, fiber.StatusOK, payload)
}

// Created sends 201 Created with payload.
func Created(c *fiber.Ctx, payload interface{}) error {
	return JSON(c, fiber.StatusCreated, payload)
}

// ErrorMessage sends error status with Hasura-compatible JSON { "message": msg }.
func ErrorMessage(c *fiber.Ctx, status int, msg string) error {
	return JSON(c, status, ErrorResponse{Message: msg})
}

// BadRequest sends 400 Bad Request.
func BadRequest(c *fiber.Ctx, msg string) error {
	return ErrorMessage(c, fiber.StatusBadRequest, msg)
}

// Unauthorized sends 401 Unauthorized.
func Unauthorized(c *fiber.Ctx, msg string) error {
	return ErrorMessage(c, fiber.StatusUnauthorized, msg)
}

// Forbidden sends 403 Forbidden.
func Forbidden(c *fiber.Ctx, msg string) error {
	return ErrorMessage(c, fiber.StatusForbidden, msg)
}

// NotFound sends 404 Not Found.
func NotFound(c *fiber.Ctx, msg string) error {
	return ErrorMessage(c, fiber.StatusNotFound, msg)
}

// Conflict sends 409 Conflict.
func Conflict(c *fiber.Ctx, msg string) error {
	return ErrorMessage(c, fiber.StatusConflict, msg)
}

// ServiceUnavailable sends 503 Service Unavailable.
func ServiceUnavailable(c *fiber.Ctx, msg string) error {
	return ErrorMessage(c, fiber.StatusServiceUnavailable, msg)
}

// Internal sends 500 Internal Server Error.
func Internal(c *fiber.Ctx, msg string) error {
	return ErrorMessage(c, fiber.StatusInternalServerError, msg)
}
