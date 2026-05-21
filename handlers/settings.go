package handlers

import (
	"super40-backend/db"
	"super40-backend/models"

	"github.com/gofiber/fiber/v2"
)

func GetSettings(c *fiber.Ctx) error {
	var settings []models.SystemSetting
	if err := db.DB.Find(&settings).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch settings"})
	}

	settingsMap := make(map[string]string)
	for _, s := range settings {
		settingsMap[s.Key] = s.Value
	}

	return c.JSON(settingsMap)
}

func UpdateSetting(c *fiber.Ctx) error {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Key == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Key is required"})
	}

	var setting models.SystemSetting
	err := db.DB.Where("key = ?", req.Key).First(&setting).Error
	if err != nil {
		// Create new setting if it does not exist
		setting.Key = req.Key
		setting.Value = req.Value
		if err := db.DB.Create(&setting).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create setting"})
		}
	} else {
		// Update existing setting
		setting.Value = req.Value
		if err := db.DB.Save(&setting).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update setting"})
		}
	}

	return c.JSON(fiber.Map{"message": "Setting updated successfully", "key": req.Key, "value": req.Value})
}
