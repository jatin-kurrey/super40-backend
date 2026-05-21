package handlers

import (
	"os"
	"time"

	"super40-backend/db"
	"super40-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(username string, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super40_highly_secret_key_2026_change_me"
	}

	return token.SignedString([]byte(jwtSecret))
}

func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	var admin models.Admin
	if result := db.DB.Where("username = ?", req.Username).First(&admin); result.Error != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !CheckPasswordHash(req.Password, admin.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := GenerateJWT(admin.Username, admin.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{"token": token, "user": admin})
}

func CreateInitialAdmin(c *fiber.Ctx) error {
	var count int64
	db.DB.Model(&models.Admin{}).Count(&count)
	if count > 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Admin already exists"})
	}

	var admin models.Admin
	if err := c.BodyParser(&admin); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if admin.Username == "" {
		admin.Username = "admin"
	}

	password := "super40_admin_2026"
	hash, _ := HashPassword(password)
	admin.PasswordHash = hash
	admin.Role = "SUPER_ADMIN"

	if result := db.DB.Create(&admin); result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Initial admin created", "username": admin.Username, "password": password})
}
