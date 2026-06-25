package middleware

import (
	"strings"

	"pet-link/internal/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

const (
	LocalUserIDKey    = "userID"
	LocalUserEmailKey = "userEmail"
)

type TokenParser interface {
	Parse(tokenString string) (jwt.Claims, error)
}

func Auth(tokens TokenParser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerToken(c.Get("Authorization"))
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or invalid authorization header",
			})
		}
		claims, err := tokens.Parse(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}
		c.Locals(LocalUserIDKey, claims.UserID)
		c.Locals(LocalUserEmailKey, claims.Email)
		return c.Next()
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func UserID(c *fiber.Ctx) (string, bool) {
	userID, ok := c.Locals(LocalUserIDKey).(string)
	return userID, ok && userID != ""
}
