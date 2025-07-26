package models

import (
	"time"
)

// User represents a user in the system, mapped to a PostgreSQL table.
type User struct {
	ID uint `gorm:"primarykey"`

	UserID   string `json:"user_id,omitempty" gorm:"uniqueIndex"` // Unique index for user_id
	Role     string `json:"role,omitempty" validate:"required,oneof=STD TCH ADMIN SAMA"`
	Email    string `json:"email,omitempty" gorm:"uniqueIndex" validate:"required,email"` // Unique index for email
	Password string `json:"-"`

	Phone             string  `json:"phone,omitempty"`
	Firstname         string  `json:"firstname,omitempty" validate:"required"`
	Lastname          string  `json:"lastname,omitempty" validate:"required"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
	IsVerified        bool    `json:"-"`
	IsActive          bool    `json:"is_active,omitempty"`

	Activities []*Activity `json:"activities,omitempty" gorm:"many2many:activity_exclusive_student_ids"`

	SchoolID  uint   `json:"school_id,omitempty" validate:"required"`
	Classroom string `json:"classroom_id,omitempty"`
	Number    uint   `json:"number,omitempty" validate:"required,number"`
	Status    string `json:"status,omitempty"`
	Language  string `json:"language,omitempty"`

	School School `json:"school" gorm:"foreignKey:SchoolID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
