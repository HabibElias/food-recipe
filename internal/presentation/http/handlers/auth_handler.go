package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/application/services"
	"go/food-recipe/internal/presentation/response"
)

// AuthHandler handles Hasura Action endpoints for authentication.
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler constructs a new AuthHandler.
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles user login Hasura action.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var payload dto.LoginActionPayloadDTO
	if err := c.BodyParser(&payload); err != nil {
		return response.BadRequest(c, "invalid payload")
	}

	res, err := h.authService.Login(c.Context(), payload.Input.Input)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}

// Signup handles user registration Hasura action.
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var payload dto.SignupActionPayloadDTO
	if err := c.BodyParser(&payload); err != nil {
		return response.BadRequest(c, "invalid payload")
	}

	res, err := h.authService.Signup(c.Context(), payload.Input.Input)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}

// Me retrieves user profile from token passed via action payload or Authorization header.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	var payload dto.MeActionPayloadDTO
	parseErr := c.BodyParser(&payload)

	token := strings.TrimSpace(payload.Input.Input.Token)
	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				token = strings.TrimSpace(parts[1])
			} else if len(parts) == 1 {
				token = strings.TrimSpace(parts[0])
			}
		}
	}

	if token == "" {
		if parseErr != nil {
			return response.BadRequest(c, "invalid payload")
		}
		return response.BadRequest(c, "token is required")
	}

	res, err := h.authService.Me(c.Context(), token)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}
