package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Record represents an activity record, mapped to a PostgreSQL table.
type Record struct {
	ID uint `gorm:"primarykey"`

	ActivityID uint                   `json:"activity_id,omitempty" validate:"required"`
	Data       map[string]interface{} `json:"data" gorm:"serializer:json"`
	Advise     string                 `json:"advise,omitempty"` // Advise might be optional

	// Foreign keys to other models
	StudentID uint `json:"student_id,omitempty" gorm:"index" validate:"required,gt=0"` // Index for faster lookups
	TeacherID uint `json:"teacher_id,omitempty" gorm:"index" validate:"required,gt=0"` // Index for faster lookups

	SchoolYear int `json:"school_year,omitempty" validate:"required,gt=0"`
	Semester   int `json:"semester,omitempty" validate:"required,gt=0"`

	Amount int `json:"amount"`

	StatusLogs StatusLogs `json:"status_logs,omitempty" gorm:"type:jsonb"`
	Status     string     `json:"status,omitempty" validate:"required,oneof=CREATED SENDED APPROVED REJECTED"`

	Activity Activity `json:"activity"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// StatusUpdateTime represents a single status update event.
type StatusHistory struct {
	Status     string    `json:"status"`
	UpdateTime time.Time `json:"update_time"`
}

// StatusUpdates is a custom type for handling []StatusUpdateTime as JSONB.
type StatusLogs []StatusHistory

// Value implements the driver.Valuer interface for StatusUpdates.
func (s StatusLogs) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]StatusLogs{}) // Return empty array for consistency
	}
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal StatusUpdates to JSON: %w", err)
	}
	return jsonBytes, nil
}

// Scan implements the sql.Scanner interface for StatusUpdates.
func (s *StatusLogs) Scan(value interface{}) error {
	if value == nil {
		*s = make(StatusLogs, 0) // Initialize to an empty slice if DB value is NULL
		return nil
	}

	var jsonBytes []byte
	switch v := value.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for StatusUpdates: %T", value)
	}

	if len(jsonBytes) == 0 {
		*s = make(StatusLogs, 0) // Handle empty JSON as empty slice
		return nil
	}

	// Ensure the slice is initialized before unmarshaling
	if *s == nil {
		*s = make(StatusLogs, 0)
	}

	err := json.Unmarshal(jsonBytes, s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to StatusUpdates: %w", err)
	}
	return nil
}

// TableName specifies the table name for the Record model.
func (Record) TableName() string {
	return "records"
}

// STATUS_ENUM defines the allowed values for the 'Status' field.
var STATUS_ENUM = []string{"CREATED", "SENDED", "APPROVED", "REJECTED"}
