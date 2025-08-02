package models

import (
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

	SchoolYear int `json:"school_year" validate:"required,gt=0"` // School year must be positive
	Semester   int `json:"semester" validate:"required,gt=0"`    // Semester must be positive

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
