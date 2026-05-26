package main

import (
	"log"
	"super40-backend/db"
	"super40-backend/models"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	db.ConnectDB()

	defaultSettings := []models.SystemSetting{
		{Key: "super40_exam_year", Value: "2024"},
		{Key: "super40_application_start", Value: "Jan 15, 2024"},
		{Key: "super40_last_date", Value: "Mar 30, 2024"},
		{Key: "super40_admit_card", Value: "Apr 15, 2024"},
		{Key: "super40_exam_date", Value: "Apr 28, 2024"},
		{Key: "super40_results_date", Value: "May 15, 2024"},
		{Key: "super40_seats", Value: "40"},
		{Key: "super40_acceptance_rate", Value: "2%"},
		{Key: "super40_placement_record", Value: "100%"},
		{Key: "super40_features", Value: "40 Seats Only - Elite Program|Merit-based Scholarships|1:10 Faculty-Student Ratio|100% Placement Guarantee|Industry-focused Curriculum|Accelerated Learning Path"},
		{Key: "super40_eligibility", Value: "Minimum 85% in 12th Grade|Mathematics & Physics compulsory|Age limit: 17-20 years|Valid JEE/CET score accepted"},
		{Key: "super40_brochure_url", Value: ""},
		{Key: "super40_total_marks", Value: "180"},
		{Key: "super40_duration_hours", Value: "3"},
		{Key: "super40_question_type", Value: "MCQ"},
	}

	for _, setting := range defaultSettings {
		var existing models.SystemSetting
		if err := db.DB.Where("key = ?", setting.Key).First(&existing).Error; err != nil {
			db.DB.Create(&setting)
			log.Printf("Created setting: %s = %s\n", setting.Key, setting.Value)
		} else {
			db.DB.Model(&existing).Update("value", setting.Value)
			log.Printf("Updated setting: %s = %s\n", setting.Key, setting.Value)
		}
	}

	log.Println("Super 40 system settings seeded successfully!")
}
