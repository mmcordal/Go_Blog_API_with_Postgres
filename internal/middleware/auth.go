package middleware

import (
	"cleanArch_with_postgres/internal/infrastructure/config"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid authorization header",
			})
		}
		if strings.HasPrefix(authHeader, "Bearer ") {
			authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// jwtSecret := []byte(os.Getenv("JWT_SECRET")) // configden çek; gereksiz yere Getenv kullanma
		jwtSecret := []byte(config.Get().Secret.JWTSecret)
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		if token == nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		if v, ok := claims["username"].(string); ok && v != "" {
			c.Locals("username", v)
		}

		if v, ok := claims["role"].(string); ok && v != "" {
			c.Locals("role", v)
		}

		// user_id (MapClaims sayısal değerleri float64 getiriyo)
		if v, ok := claims["user_id"]; ok {
			switch id := v.(type) {
			case float64:
				uid := uint(id)
				c.Locals("user_id", uid)
				c.Locals("userID", uid)
				c.Locals("userId", uid)

			case int64:
				uid := uint(id)
				c.Locals("user_id", uid)
				c.Locals("userID", uid)
				c.Locals("userId", uid)

			case int:
				uid := uint(id)
				c.Locals("user_id", uid)
				c.Locals("userID", uid)
				c.Locals("userId", uid)

			}
		}

		return c.Next()
	}
}
