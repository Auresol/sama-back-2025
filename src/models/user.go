package models

import (
	"time"
)

// User represents a user in the system, mapped to a PostgreSQL table.
type User struct {
	ID uint `gorm:"primarykey"`

	UserID   string `json:"user_id,omitempty" gorm:"column:user_id;uniqueIndex"` // Unique index for user_id
	Role     string `json:"role,omitempty" gorm:"column:role" validate:"required"`
	Email    string `json:"email,omitempty" gorm:"column:email;uniqueIndex" validate:"required,email"` // Unique index for email
	Password string `json:"-" gorm:"column:password"`

	Phone             string  `json:"phone,omitempty" gorm:"column:phone"`
	Firstname         string  `json:"firstname,omitempty" gorm:"column:firstname" validate:"required"`
	Lastname          string  `json:"lastname,omitempty" gorm:"column:lastname" validate:"required"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty" gorm:"column:profile_picture_url"`
	IsActive          bool    `json:"is_active,omitempty" gorm:"column:is_active"`

	Classroom string `json:"classroom,omitempty" gorm:"column:classroom"`
	Number    int    `json:"number,omitempty" gorm:"column:number" validate:"required,number"`
	SchoolID  uint   `json:"school_id,omitempty" gorm:"column:school_id"`
	Status    string `json:"status,omitempty" gorm:"column:status"`
	Language  string `json:"language,omitempty" gorm:"column:language"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
