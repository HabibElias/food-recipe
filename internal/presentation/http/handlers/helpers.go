package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// resolveSessionUserID extracts user ID from Hasura session variables or fiber context locals.
func resolveSessionUserID(sessionVars map[string]interface{}, c *fiber.Ctx) string {
	if sessionVars != nil {
		for _, key := range []string{"x-hasura-user-id", "hasura_user_id", "HasuraUserID", "user_id"} {
			if v, ok := sessionVars[key]; ok {
				if s, ok := v.(string); ok {
					s = strings.TrimSpace(s)
					if s != "" {
						return s
					}
				}
			}
		}
	}

	if c != nil {
		if v, ok := c.Locals("user_id").(string); ok {
			s := strings.TrimSpace(v)
			if s != "" {
				return s
			}
		}
	}

	return ""
}
