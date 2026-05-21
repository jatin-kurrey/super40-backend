package handlers

import (
	"time"

	"super40-backend/db"
	"super40-backend/models"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

func CreateApplication(c *fiber.Ctx) error {
	app := new(models.Application)
	if err := c.BodyParser(app); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if app.FormType == "" {
		app.FormType = "super40"
	}

	if result := db.DB.Create(&app); result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
	}

	return c.Status(201).JSON(app)
}

func GetApplications(c *fiber.Ctx) error {
	type ApplicationWithScore struct {
		ID         uuid.UUID      `json:"id"`
		FormType   string         `json:"form_type"`
		Name       string         `json:"name"`
		Email      string         `json:"email"`
		Phone      string         `json:"phone"`
		Data       datatypes.JSON `json:"data"`
		Status     string         `json:"status"`
		Score      *int           `json:"score"`
		ResponseID *uuid.UUID     `json:"response_id"`
		CreatedAt  time.Time      `json:"created_at"`
	}
	var results []ApplicationWithScore

	formType := c.Query("type")
	status := c.Query("status")

	query := db.DB.Table("applications").
		Select("applications.*, r.max_score as score, r.best_response_id as response_id").
		Joins("LEFT JOIN (" +
			"SELECT DISTINCT ON (LOWER(student_id)) LOWER(student_id) as student_id_lower, score as max_score, id as best_response_id " +
			"FROM exam_responses " +
			"ORDER BY LOWER(student_id), score DESC" +
			") r ON r.student_id_lower = LOWER(applications.email)")

	if formType != "" {
		query = query.Where("applications.form_type = ?", formType)
	}
	if status != "" {
		query = query.Where("applications.status = ?", status)
	}

	query = query.Order("applications.created_at DESC")

	if result := query.Scan(&results); result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
	}

	return c.JSON(results)
}

func UpdateApplicationStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	type StatusUpdate struct {
		Status string `json:"status"`
	}
	var update StatusUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if result := db.DB.Model(&models.Application{}).Where("id = ?", id).Update("status", update.Status); result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
	}

	return c.JSON(fiber.Map{"message": "Status updated successfully"})
}
