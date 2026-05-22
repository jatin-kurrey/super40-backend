package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(c *fiber.Ctx) error {
	if c.Method() == "OPTIONS" {
		return c.Next()
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Missing authorization header"})
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "super40_highly_secret_key_2026_change_me"
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)

	c.Locals("username", username)
	c.Locals("role", role)

	return c.Next()
}
