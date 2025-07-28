package models

import (
	"time"

	"gorm.io/gorm"
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

	SchoolID  uint    `json:"school_id" validate:"required"`
	Classroom *string `json:"classroom,omitempty" gorm:"-:all" validate:"classroomregex"`
	Number    *uint   `json:"number,omitempty" validate:"gt=0"`

	ClassroomID    *uint       `json:"-"`
	ClassroomModel *Classroom  `json:"-" gorm:"foreignKey:ClassroomID"`
	School         School      `json:"school" gorm:"foreignKey:SchoolID"`
	Activities     []*Activity `json:"-" gorm:"many2many:activity_exclusive_student_ids"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {
	// Ensure Classroom is loaded before attempting to flatten
	// This requires preloading Classroom in your repository's Get methods.
	if u.ClassroomModel != nil {
		u.Classroom = &u.ClassroomModel.Classroom
	}
	return nil
}

var ROLE = []string{"STD", "TCH", "ADMIN", "SAMA_CREW"}
