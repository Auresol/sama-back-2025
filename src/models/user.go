package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system, mapped to a PostgreSQL table.
type User struct {
	ID uint `json:"id" gorm:"primarykey"`

	StudentID         string  `json:"student_id,omitempty"`
	Role              string  `json:"role" validate:"required,oneof=STD TCH ADMIN SAMA"`
	Email             string  `json:"email" gorm:"uniqueIndex" validate:"required,email"` // Unique index for email
	Password          string  `json:"-"`
	Phone             string  `json:"phone,omitempty"`
	Firstname         string  `json:"firstname" validate:"required"`
	Lastname          string  `json:"lastname" validate:"required"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
	Language          string  `json:"language" validate:"required"`

	SchoolID          uint    `json:"school_id" validate:"required"`
	Classroom         *string `json:"classroom,omitempty"`
	Number            *uint   `json:"number,omitempty" validate:"gt=0"`
	BookmarkedUserIDs []uint  `json:"bookmarked_user_ids,omitempty" gorm:"-:all"`

	ClassroomID    *uint       `json:"-"`
	ClassroomModel *Classroom  `json:"-" gorm:"foreignKey:ClassroomID"`
	School         School      `json:"school" gorm:"foreignKey:SchoolID"`
	Activities     []*Activity `json:"-" gorm:"many2many:activity_exclusive_student_ids"`
	BookmarkedUser []*User     `json:"-" gorm:"many2many:user_bookmark"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggertype:"string"`
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
