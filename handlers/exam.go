package handlers

import (
	"encoding/json"
	"strings"

	"super40-backend/db"
	"super40-backend/models"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
)

// Exam Management (Admin)

func CreateExam(c *fiber.Ctx) error {
	exam := new(models.Exam)
	if err := c.BodyParser(exam); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := db.DB.Create(exam).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create exam"})
	}

	return c.Status(201).JSON(exam)
}

func GetExamsAdmin(c *fiber.Ctx) error {
	type ExamWithStats struct {
		models.Exam
		Appearances int     `json:"appearances"`
		AvgScore    float64 `json:"avg_score"`
	}
	var exams []models.Exam
	if err := db.DB.Preload("Questions").Order("created_at desc").Find(&exams).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch exams"})
	}

	var stats []struct {
		ExamID uuid.UUID
		Count  int
		Avg    float64
	}
	db.DB.Model(&models.ExamResponse{}).
		Select("exam_id, count(*) as count, coalesce(avg(score), 0) as avg").
		Group("exam_id").
		Scan(&stats)

	statsMap := make(map[uuid.UUID]struct {
		Count int
		Avg   float64
	})
	for _, s := range stats {
		statsMap[s.ExamID] = struct {
			Count int
			Avg   float64
		}{Count: s.Count, Avg: s.Avg}
	}

	var results []ExamWithStats
	for _, exam := range exams {
		st := statsMap[exam.ID]
		results = append(results, ExamWithStats{
			Exam:        exam,
			Appearances: st.Count,
			AvgScore:    st.Avg,
		})
	}

	return c.JSON(results)
}

func UpdateExam(c *fiber.Ctx) error {
	id := c.Params("id")
	exam := new(models.Exam)
	if err := c.BodyParser(exam); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	uuidID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid exam ID"})
	}
	exam.ID = uuidID

	tx := db.DB.Begin()

	// Update the main exam details
	if err := tx.Model(&models.Exam{}).Where("id = ?", id).Updates(exam).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update exam"})
	}

	// For questions, we'll replace the existing ones to simplify synchronization
	if len(exam.Questions) > 0 {
		// Delete existing questions
		if err := tx.Where("exam_id = ?", id).Delete(&models.Question{}).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to sync questions"})
		}
		// Insert new ones
		for i := range exam.Questions {
			exam.Questions[i].ExamID = exam.ID
		}
		if err := tx.Create(&exam.Questions).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save new questions"})
		}
	}

	tx.Commit()
	return c.JSON(fiber.Map{"message": "Exam updated successfully"})
}

func DeleteExam(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.DB.Delete(&models.Exam{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete exam"})
	}
	return c.JSON(fiber.Map{"message": "Exam deleted successfully"})
}

// Exam Interaction (Public/Student)

func GetActiveExams(c *fiber.Ctx) error {
	var exams []models.Exam
	if err := db.DB.Where("is_active = ?", true).Order("start_time asc").Find(&exams).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch exams"})
	}
	return c.JSON(exams)
}

func GetExamDetails(c *fiber.Ctx) error {
	id := c.Params("id")
	var exam models.Exam
	if err := db.DB.Preload("Questions").First(&exam, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found"})
	}
	return c.JSON(exam)
}

func SubmitExamResponse(c *fiber.Ctx) error {
	response := new(models.ExamResponse)
	if err := c.BodyParser(response); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var exam models.Exam
	if err := db.DB.Preload("Questions").First(&exam, "id = ?", response.ExamID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found"})
	}

	var studentResponses map[string]string
	json.Unmarshal(response.Responses, &studentResponses)

	score := 0
	for _, q := range exam.Questions {
		trimmedStudentAnswer := strings.TrimSpace(studentResponses[q.ID.String()])
		trimmedCorrectAnswer := strings.TrimSpace(q.CorrectAnswer)

		if strings.EqualFold(trimmedStudentAnswer, trimmedCorrectAnswer) {
			score += q.Points
		}
	}

	response.Score = score
	response.Submitted = true

	if err := db.DB.Create(&response).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to submit response"})
	}

	return c.Status(201).JSON(response)
}

func GetExamResponses(c *fiber.Ctx) error {
	examID := c.Params("id")
	var responses []models.ExamResponse
	if err := db.DB.Where("exam_id = ?", examID).Order("score desc").Find(&responses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch responses"})
	}
	return c.JSON(responses)
}

func GetStudentResults(c *fiber.Ctx) error {
	email := c.Query("email")
	phone := c.Query("phone")

	if email == "" || phone == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and Phone are required"})
	}

	// Verify application exists first (authentication)
	var app models.Application
	if err := db.DB.Where("email = ? AND phone = ?", email, phone).First(&app).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "No registration found with these credentials. Please register for Super 40 first."})
	}

	var responses []models.ExamResponse
	if err := db.DB.Where("student_id = ?", email).Order("created_at desc").Find(&responses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch results"})
	}

	type ResultWithTitle struct {
		models.ExamResponse
		ExamTitle string `json:"exam_title"`
	}

	results := make([]ResultWithTitle, 0)
	for _, resp := range responses {
		var exam models.Exam
		ExamTitle := "Super 40 Evaluation"
		if err := db.DB.Select("title").First(&exam, "id = ?", resp.ExamID).Error; err == nil {
			ExamTitle = exam.Title
		}

		results = append(results, ResultWithTitle{
			ExamResponse: resp,
			ExamTitle:    ExamTitle,
		})
	}

	return c.JSON(results)
}

func GetDetailedResponse(c *fiber.Ctx) error {
	id := c.Params("id")
	var response models.ExamResponse
	if err := db.DB.First(&response, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Response not found: " + err.Error()})
	}

	var exam models.Exam
	if err := db.DB.Preload("Questions").First(&exam, "id = ?", response.ExamID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Exam not found: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"response": response,
		"exam":     exam,
	})
}

func ActivateExam(c *fiber.Ctx) error {
	id := c.Params("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid exam ID"})
	}

	tx := db.DB.Begin()

	// Deactivate all exams first
	if err := tx.Model(&models.Exam{}).Where("1 = 1").Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to deactivate other exams"})
	}

	// Activate this specific exam
	if err := tx.Model(&models.Exam{}).Where("id = ?", uuidID).Update("is_active", true).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to activate exam"})
	}

	tx.Commit()
	return c.JSON(fiber.Map{"message": "Exam activated successfully"})
}
