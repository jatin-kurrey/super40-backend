package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Admin struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"default:'EDITOR'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Application struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FormType  string         `gorm:"not null" json:"form_type"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"index;not null" json:"email"`
	Phone     string         `gorm:"not null" json:"phone"`
	Data      datatypes.JSON `json:"data"`
	Status    string         `gorm:"default:'pending'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Exam struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title                 string     `gorm:"not null" json:"title"`
	Description           string     `json:"description"`
	Duration              int        `json:"duration"` // in minutes
	NegativeMarking       float64    `gorm:"default:0" json:"negative_marking"`
	ShuffleQuestions      bool       `gorm:"default:false" json:"shuffle_questions"`
	BrowserLockdown       bool       `gorm:"default:false" json:"browser_lockdown"`
	ShowResultImmediately bool       `gorm:"default:false" json:"show_result_immediately"`
	StartTime             time.Time  `json:"start_time"`
	EndTime               time.Time  `json:"end_time"`
	IsActive              bool       `gorm:"default:false" json:"is_active"`
	Questions             []Question `gorm:"foreignKey:ExamID;constraint:OnDelete:CASCADE" json:"questions"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type Question struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExamID        uuid.UUID      `gorm:"type:uuid;not null" json:"exam_id"`
	Text          string         `gorm:"not null" json:"text"`
	Type          string         `gorm:"not null" json:"type"` // MCQ, INTEGER
	Options       datatypes.JSON `json:"options"`             // Array of strings
	CorrectAnswer string         `gorm:"not null" json:"correct_answer"`
	Points        int            `gorm:"default:1" json:"points"`
	ImageURL      string         `json:"image_url"`
	Subject       string         `json:"subject"`   // e.g. Mathematics, Physics
	Difficulty    string         `json:"difficulty"` // e.g. Easy, Medium, Hard
	CreatedAt     time.Time      `json:"created_at"`
}

type ExamResponse struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExamID    uuid.UUID      `gorm:"type:uuid;not null" json:"exam_id"`
	StudentID string         `gorm:"index;not null" json:"student_id"`
	Name      string         `json:"name"`
	Responses datatypes.JSON `json:"responses"` // question_id -> answer
	Score     int            `json:"score"`
	Submitted bool           `gorm:"default:false" json:"submitted"`
	CreatedAt time.Time      `json:"created_at"`
}

type SystemSetting struct {
	Key       string    `gorm:"primaryKey" json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
