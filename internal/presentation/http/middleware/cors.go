package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware configures CORS for incoming requests.
func CORSMiddleware(allowedOrigins string) fiber.Handler {
	allowCredentials := true
	if allowedOrigins == "" || allowedOrigins == "*" {
		allowedOrigins = "*"
		allowCredentials = false
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowCredentials: allowCredentials,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
		ExposeHeaders:    "Link,Authorization",
		MaxAge:           300,
	})
}
