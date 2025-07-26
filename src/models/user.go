package models

import (
	"time"
)

// User represents a user in the system, mapped to a PostgreSQL table.
type User struct {
	ID uint `gorm:"primarykey" validate:"required"`

	StudentID         string  `json:"student_id,omitempty"` // Unique index for user_id
	Role              string  `json:"role" validate:"required,oneof=STD TCH ADMIN SAMA"`
	Email             string  `json:"email" gorm:"uniqueIndex" validate:"required,email"` // Unique index for email
	Password          string  `json:"-"`
	Phone             string  `json:"phone,omitempty"`
	Firstname         string  `json:"firstname" validate:"required"`
	Lastname          string  `json:"lastname" validate:"required"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
	Language          string  `json:"language" validate:"required"`

	SchoolID  uint   `json:"school_id" validate:"required"`
	Classroom string `json:"classroom,omitempty" validate:"classroomregex"`
	Number    uint   `json:"number,omitempty" validate:"gt=0"`

	School     School      `json:"school,omitempty" gorm:"foreignKey:SchoolID"`
	Activities []*Activity `json:"-" gorm:"many2many:activity_exclusive_student_ids"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
