package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware configures CORS for incoming requests.
func CORSMiddleware(allowedOrigins string) fiber.Handler {
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
		ExposeHeaders:    "Link,Authorization",
		AllowCredentials: true,
		MaxAge:           300,
	})
}
