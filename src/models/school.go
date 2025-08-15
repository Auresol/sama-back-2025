package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// School represents a school entity, mapped to a PostgreSQL table.
type School struct {
	ID uint `json:"id" gorm:"primarykey"`

	ThaiName                string    `json:"thai_name" validate:"required"`
	EnglishName             string    `json:"english_name" validate:"required"`
	ShortName               string    `json:"short_name" gorm:"uniqueIndex" validate:"required"` // Added unique index for short_name
	SchoolLogoUrl           *string   `json:"school_logo_url"`
	Email                   *string   `json:"email,omitempty" validate:"email"`
	Location                *string   `json:"location,omitempty"`
	Phone                   *string   `json:"phone,omitempty" validate:"e164"` // e164 for phone number validation
	DefaultActivityDeadline time.Time `json:"default_activity_deadline" validate:"required"`

	Classrooms []string `json:"classrooms" gorm:"-:all" validate:"required"`

	SchoolYear            uint             `json:"school_year" validate:"required,gt=0"` // School year must be positive
	Semester              uint             `json:"semester" validate:"required,gt=0"`    // Semester must be positive\
	AvaliableSemesterList SemesterYearList `json:"avaliable_semester_list" gorm:"serializer:json"`

	ClassroomObjects []Classroom `json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggertype:"string"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (School) TableName() string {
	return "schools"
}

// AfterFind is a GORM callback that runs after a record is found.
// It populates the `Classrooms []string` field from the `ClassroomList` association.
func (s *School) AfterFind(tx *gorm.DB) (err error) {
	// Ensure ClassroomList is loaded before attempting to flatten
	// This requires preloading ClassroomList in your repository's Get methods.
	for _, obj := range s.ClassroomObjects {
		s.Classrooms = append(s.Classrooms, obj.Classroom)
	}
	return nil
}

// SemesterYearList is a slice of slices, representing pairs of [semester, year].
type SemesterYearList []string

// Value converts the SemesterYearStringList to a JSON byte slice for storage.
func (s SemesterYearList) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SemesterYearStringList: %w", err)
	}
	return jsonBytes, nil
}

// Scan converts a JSON byte slice from the database back into a SemesterYearStringList.
func (s *SemesterYearList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("SemesterYearStringList Scan: unsupported value type: %T", value)
	}
	if len(bytes) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(bytes, s)
}
