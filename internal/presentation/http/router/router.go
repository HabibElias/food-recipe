package router

import (
	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/presentation/http/handlers"
	"go/food-recipe/internal/presentation/http/middleware"
)

// RouterParams encapsulates dependencies required to mount all HTTP routes.
type RouterParams struct {
	App                 *fiber.App
	AuthHandler         *handlers.AuthHandler
	PaymentHandler      *handlers.PaymentHandler
	MediaHandler        *handlers.MediaHandler
	NotificationHandler *handlers.NotificationHandler
	HealthHandler       *handlers.HealthHandler
	ActionSecret        string
}

// SetupRoutes mounts all API endpoints under /v1.
func SetupRoutes(p RouterParams) {
	v1 := p.App.Group("/v1")
	webhook := middleware.HasuraWebhookMiddleware(p.ActionSecret)

	// Health
	if p.HealthHandler != nil {
		v1.Get("/healthz", p.HealthHandler.Healthz)
		v1.Get("/err", p.HealthHandler.Err)
	}

	// Auth Hasura Actions
	if p.AuthHandler != nil {
		v1.Post("/login", webhook, p.AuthHandler.Login)
		v1.Post("/signup", webhook, p.AuthHandler.Signup)
		v1.Post("/me", webhook, p.AuthHandler.Me)
	}

	// Payment
	if p.PaymentHandler != nil {
		v1.Post("/payment", webhook, p.PaymentHandler.InitializePayment)
		v1.Get("/payment/verify", p.PaymentHandler.VerifyPayment)
	}

	// Media Hasura Actions
	if p.MediaHandler != nil {
		v1.Post("/upload-recipe-images", webhook, p.MediaHandler.UploadRecipeImages)
		v1.Post("/upload-user-image", webhook, p.MediaHandler.UploadUserImage)
	}

	// Notification Webhook
	if p.NotificationHandler != nil {
		v1.Post("/like-notification", webhook, p.NotificationHandler.LikeNotification)
	}
}
