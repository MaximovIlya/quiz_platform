package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/auth"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

const (
	CtxUserID = "userID"
	CtxRole   = "role"
)

func Auth(manager *auth.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return fiber.ErrUnauthorized
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.ErrUnauthorized
		}

		claims, err := manager.Parse(parts[1])
		if err != nil {
			return fiber.ErrUnauthorized
		}

		if claims.Type != auth.AccessToken {
			return fiber.ErrUnauthorized
		}

		c.Locals(CtxUserID, claims.UserID)
		c.Locals(CtxRole, claims.Role)

		return c.Next()
	}
}

func RequireRole(role db.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Locals(CtxRole) != role {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}
