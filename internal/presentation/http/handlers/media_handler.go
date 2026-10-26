package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/application/services"
	"go/food-recipe/internal/presentation/response"
)

// MediaHandler handles media upload Hasura actions.
type MediaHandler struct {
	mediaService *services.MediaService
}

// NewMediaHandler constructs a new MediaHandler.
func NewMediaHandler(mediaService *services.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// UploadRecipeImages handles recipe images upload Hasura action.
func (h *MediaHandler) UploadRecipeImages(c *fiber.Ctx) error {
	var payload dto.UploadRecipeImagesActionPayloadDTO
	if err := c.BodyParser(&payload); err != nil {
		return response.BadRequest(c, "invalid payload")
	}

	input := payload.Input.Input
	if err := input.Validate(); err != nil {
		return response.HandleDomainError(c, err)
	}

	res, err := h.mediaService.UploadRecipeImages(c.Context(), input)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}

// UploadUserImage handles user avatar upload Hasura action.
func (h *MediaHandler) UploadUserImage(c *fiber.Ctx) error {
	var payload dto.UploadUserImageActionPayloadDTO
	if err := c.BodyParser(&payload); err != nil {
		return response.BadRequest(c, "invalid payload")
	}

	sessionUserID := resolveSessionUserID(payload.SessionVariables, c)

	input := payload.Input.Input
	if input.UserID == "" {
		input.UserID = sessionUserID
	}

	if err := input.Validate(); err != nil {
		return response.HandleDomainError(c, err)
	}

	res, err := h.mediaService.UploadUserImage(c.Context(), input, sessionUserID)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}
