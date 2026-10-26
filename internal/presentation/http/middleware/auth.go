package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/infrastructure/security"
	"go/food-recipe/internal/presentation/response"
)

// AuthMiddleware validates JWT Bearer tokens and populates fiber context locals.
func AuthMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "missing authorization header")
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return response.Unauthorized(c, "invalid authorization header format")
		}

		claims, err := security.VerifyHasuraJWT(parts[1], secret)
		if err != nil {
			return response.Unauthorized(c, err.Error())
		}

		c.Locals("claims", claims)
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			c.Locals("user_id", sub)
		} else if hasuraClaims, ok := claims["https://hasura.io/jwt/claims"].(map[string]interface{}); ok {
			if uid, ok := hasuraClaims["x-hasura-user-id"].(string); ok && uid != "" {
				c.Locals("user_id", uid)
			}
		}

		return c.Next()
	}
}
