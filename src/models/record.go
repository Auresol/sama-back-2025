package models

import (
	"time"

	"gorm.io/gorm"
)

// Record represents an activity record, mapped to a PostgreSQL table.
type Record struct {
	ID uint `gorm:"primarykey"`

	ActivityID uint                   `json:"activity_id" validate:"required"`
	Data       map[string]interface{} `json:"data" gorm:"serializer:json" validate:"required"`
	Advise     *string                `json:"advise,omitempty"` // Advise might be optional

	// Foreign keys to other models
	StudentID uint  `json:"student_id" gorm:"index" validate:"required,gt=0"`  // Index for faster lookups
	TeacherID *uint `json:"teacher_id,omitempty" gorm:"index" validate:"gt=0"` // Index for faster lookups

	SchoolYear int `json:"school_year" validate:"required,gt=0"`
	Semester   int `json:"semester" validate:"required,gt=0"`

	Amount int `json:"amount" validate:"required"`

	StatusLogs StatusLogs `json:"status_logs" gorm:"serializer:json" validate:"required"`
	Status     string     `json:"status" validate:"required,oneof=CREATED SENDED APPROVED REJECTED"`

	Activity Activity `json:"-"`
	Student  User     `json:"-"`
	Teacher  *User    `json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggertype:"string"`
}

// StatusUpdates is a custom type for handling []StatusUpdateTime as JSONB.
type StatusLogs []StatusHistory

// StatusUpdateTime represents a single status update event.
type StatusHistory struct {
	Status     string    `json:"status" validate:"required"`
	UpdateTime time.Time `json:"update_time" validate:"required"`
}

// TableName specifies the table name for the Record model.
func (Record) TableName() string {
	return "records"
}

// STATUS_ENUM defines the allowed values for the 'Status' field.
var STATUS_ENUM = []string{"CREATED", "SENDED", "APPROVED", "REJECTED"}
