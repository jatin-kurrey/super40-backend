package main

import (
	"log"
	"os"

	"super40-backend/db"
	"super40-backend/handlers"
	"super40-backend/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
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

	// Connect to Database
	db.ConnectDB()

	// Initialize Fiber app
	app := fiber.New()

	// Global Middlewares
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health Check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Super 40 API Server is running!")
	})

	// V1 API Router Group
	api := app.Group("/api/v1")

	// Authentication (Public)
	api.Post("/auth/login", handlers.Login)
	api.Post("/auth/init", handlers.CreateInitialAdmin)

	// Applications (Public)
	api.Post("/applications", handlers.CreateApplication)

	// Exams (Public)
	api.Get("/exams/active", handlers.GetActiveExams)
	api.Get("/exams/results", handlers.GetStudentResults)
	api.Get("/exams/:id", handlers.GetExamDetails)
	api.Post("/exams/submit", handlers.SubmitExamResponse)
	api.Get("/settings", handlers.GetSettings)

	// Admin (Protected)
	admin := api.Group("/admin", middleware.AuthRequired)

	// Admin Settings & Security
	admin.Put("/profile/password", handlers.ChangeSelfPassword)

	// Applications Management
	admin.Get("/applications", handlers.GetApplications)
	admin.Put("/applications/:id/status", handlers.UpdateApplicationStatus)

	// Exams Management
	admin.Post("/exams", handlers.CreateExam)
	admin.Get("/exams", handlers.GetExamsAdmin)
	admin.Put("/exams/:id", handlers.UpdateExam)
	admin.Delete("/exams/:id", handlers.DeleteExam)
	admin.Get("/exams/:id/responses", handlers.GetExamResponses)
	admin.Get("/responses/:id", handlers.GetDetailedResponse)
	admin.Post("/exams/:id/activate", handlers.ActivateExam)
	admin.Put("/settings", handlers.UpdateSetting)

	// File Upload Placeholder to prevent frontend failures
	admin.Post("/upload", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"url": "/uploads/placeholder.jpg"})
	})

	// Start Listening
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Super 40 API Server listening on port :%s", port)
	log.Fatal(app.Listen(":" + port))
}
