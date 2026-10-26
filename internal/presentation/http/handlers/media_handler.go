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

	res, err := h.mediaService.UploadRecipeImages(c.Context(), payload.Input.Input)
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

	sessionUserID := payload.GetSessionUserID()
	if sessionUserID == "" && payload.SessionVariables != nil {
		for _, key := range []string{"x-hasura-user-id", "hasura_user_id", "HasuraUserID", "user_id"} {
			if v, ok := payload.SessionVariables[key]; ok {
				if s, ok := v.(string); ok && s != "" {
					sessionUserID = s
					break
				}
			}
		}
	}
	if sessionUserID == "" {
		if v, ok := c.Locals("user_id").(string); ok {
			sessionUserID = v
		}
	}

	res, err := h.mediaService.UploadUserImage(c.Context(), payload.Input.Input, sessionUserID)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}
