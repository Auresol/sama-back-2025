package models

import (
	"time"
)

// School represents a school entity, mapped to a PostgreSQL table.
type School struct {
	ID uint `gorm:"primarykey"`

	ThaiName    string `json:"thai_name,omitempty" gorm:"column:thai_name" validate:"required"`
	EnglishName string `json:"english_name,omitempty" gorm:"column:english_name" validate:"required"`
	ShortName   string `json:"short_name,omitempty" gorm:"column:short_name;uniqueIndex" validate:"required"` // Added unique index for short_name
	Email       string `json:"email,omitempty" gorm:"column:email;uniqueIndex" validate:"required,email"`     // Added unique index for email
	Location    string `json:"location,omitempty" gorm:"column:location" validate:"required"`
	Phone       string `json:"phone,omitempty" gorm:"column:phone" validate:"required,e164"` // e164 for phone number validation

	SchoolYear int `json:"school_year" gorm:"column:school_year" validate:"required,gt=0"` // School year must be positive
	Semester   int `json:"semester" gorm:"column:semester" validate:"required,gt=0"`       // Semester must be positive

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (School) TableName() string {
	return "schools"
}
