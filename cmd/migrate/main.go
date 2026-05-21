package main

import (
	"log"
	"time"

	"super40-backend/db"
	"super40-backend/handlers"
	"super40-backend/models"

	"github.com/joho/godotenv"
	"gorm.io/datatypes"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		err = godotenv.Load("../../.env")
	}
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect database
	db.ConnectDB()

	log.Println("Running AutoMigration for Super 40 tables...")
	err = db.DB.AutoMigrate(
		&models.Admin{},
		&models.Application{},
		&models.Exam{},
		&models.Question{},
		&models.ExamResponse{},
		&models.SystemSetting{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database schemas successfully migrated")

	// Seed initial admin if none exists
	var count int64
	db.DB.Model(&models.Admin{}).Count(&count)
	if count == 0 {
		log.Println("No admin user found. Creating default admin...")
		password := "super40_admin_2026"
		hash, err := handlers.HashPassword(password)
		if err != nil {
			log.Fatalf("Failed to hash default admin password: %v", err)
		}

		admin := models.Admin{
			Username:     "admin",
			PasswordHash: hash,
			Role:         "SUPER_ADMIN",
		}
		if err := db.DB.Create(&admin).Error; err != nil {
			log.Fatalf("Failed to create default admin: %v", err)
		}
		log.Printf("Successfully created default admin user: username=%s, password=%s", admin.Username, password)
	} else {
		log.Println("Admin users already exist, seeding skipped")
	}

	// Seed sample exam if none exists
	var examCount int64
	db.DB.Model(&models.Exam{}).Count(&examCount)
	if examCount == 0 {
		log.Println("No exams found. Seeding a sample Super 40 entrance exam...")
		sampleExam := models.Exam{
			Title:                 "Super 40 Entrance Evaluation 2026",
			Description:           "Official entrance assessment for Krishna Engineering College's prestigious Super 40 coding program. The exam covers Physics, Chemistry, and Mathematics concepts with high analytical standards.",
			Duration:              60,
			NegativeMarking:       0.25,
			ShuffleQuestions:      true,
			BrowserLockdown:       false,
			ShowResultImmediately: true,
			IsActive:              true,
			StartTime:             time.Now(),
			EndTime:               time.Now().AddDate(0, 1, 0), // Active for 1 month
			Questions: []models.Question{
				{
					Text:          "What is the dimensional formula for the Universal Gravitational Constant (G)?",
					Type:          "MCQ",
					Options:       datatypes.JSON([]byte(`["[M^-1 L^3 T^-2]", "[M^1 L^3 T^-2]", "[M^-1 L^2 T^-2]", "[M^-1 L^3 T^-1]"]`)),
					CorrectAnswer: "[M^-1 L^3 T^-2]",
					Points:        4,
					Subject:       "Physics",
					Difficulty:    "Medium",
				},
				{
					Text:          "Which of the following chemical elements has the highest electronegativity value on the Pauling scale?",
					Type:          "MCQ",
					Options:       datatypes.JSON([]byte(`["Fluorine", "Chlorine", "Oxygen", "Nitrogen"]`)),
					CorrectAnswer: "Fluorine",
					Points:        4,
					Subject:       "Chemistry",
					Difficulty:    "Easy",
				},
				{
					Text:          "Evaluate the limit: lim (x -> 0) [sin(x) / x].",
					Type:          "INTEGER",
					Options:       datatypes.JSON([]byte(`[]`)),
					CorrectAnswer: "1",
					Points:        4,
					Subject:       "Mathematics",
					Difficulty:    "Medium",
				},
				{
					Text:          "A moving body of mass 2 kg possesses a velocity of 5 m/s. Calculate its total kinetic energy in Joules.",
					Type:          "INTEGER",
					Options:       datatypes.JSON([]byte(`[]`)),
					CorrectAnswer: "25",
					Points:        4,
					Subject:       "Physics",
					Difficulty:    "Easy",
				},
			},
		}

		if err := db.DB.Create(&sampleExam).Error; err != nil {
			log.Fatalf("Failed to seed sample exam: %v", err)
		}
		log.Println("Successfully seeded professional sample exam with 4 JEE-style questions!")
	} else {
		log.Println("Exams already exist, sample exam seeding skipped")
	}

	// Seed system settings if none exists
	var settingsCount int64
	db.DB.Model(&models.SystemSetting{}).Count(&settingsCount)
	if settingsCount == 0 {
		log.Println("Seeding default system settings...")
		db.DB.Create(&models.SystemSetting{Key: "direct_exam_mode", Value: "true"})
		log.Println("Successfully seeded default system settings!")
	} else {
		log.Println("System settings already exist, seeding skipped")
	}

	log.Println("Migration process completed successfully")
}
