package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/application/services"
	"go/food-recipe/internal/presentation/response"
)

// NotificationHandler handles Hasura event webhooks for notifications.
type NotificationHandler struct {
	notificationService *services.NotificationService
}

// NewNotificationHandler constructs a new NotificationHandler.
func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// LikeNotification handles recipe like event webhook from Hasura.
func (h *NotificationHandler) LikeNotification(c *fiber.Ctx) error {
	var payload dto.LikeEventPayloadDTO
	if err := c.BodyParser(&payload); err != nil {
		return response.BadRequest(c, "invalid payload")
	}

	if !strings.EqualFold(payload.Event.Op, "INSERT") {
		return response.OK(c, &dto.LikeNotificationOutputDTO{OK: true, Skipped: true})
	}

	res, err := h.notificationService.HandleLikeNotification(c.Context(), payload)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}
