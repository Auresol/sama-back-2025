package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system, mapped to a PostgreSQL table.
type User struct {
	ID uint `json:"id" gorm:"primarykey"`

	StudentUniqueID   string  `json:"student_id,omitempty"`
	Role              string  `json:"role" validate:"required,oneof=STD TCH ADMIN SAMA"`
	Email             string  `json:"email" gorm:"uniqueIndex" validate:"required,email"` // Unique index for email
	Password          string  `json:"-"`
	Phone             string  `json:"phone,omitempty"`
	Firstname         string  `json:"firstname" validate:"required"`
	Lastname          string  `json:"lastname" validate:"required"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
	Language          string  `json:"language" validate:"required"`

	SchoolID        uint    `json:"school_id" validate:"required"`
	Classroom       *string `json:"classroom,omitempty"`
	Number          *uint   `json:"number,omitempty" validate:"gt=0"`
	BookmarkUserIDs []uint  `json:"bookmark_user_ids" gorm:"-:all"`

	ClassroomID     *uint      `json:"-"`
	ClassroomObject *Classroom `json:"-" gorm:"foreignKey:ClassroomID"`
	School          School     `json:"school,omitzero"`
	Activities      []Activity `json:"-" gorm:"many2many:activity_exclusive_student_ids"`
	BookmarkUsers   []User     `json:"-" gorm:"many2many:user_bookmarks"`

	FinishedPercent uint `json:"finished_percent,omitempty" gorm:"-:all"`

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
	if u.ClassroomObject != nil {
		u.Classroom = &u.ClassroomObject.Classroom
	}

	u.BookmarkUserIDs = make([]uint, len(u.BookmarkUsers))
	for i, user := range u.BookmarkUsers {
		u.BookmarkUserIDs[i] = user.ID
	}
	return nil
}

var ROLE = []string{"STD", "TCH", "ADMIN", "SAMA"}

type UserWithFinishedPercent struct {
	User
	FinishedPercent float32 `json:"finished_percent" gorm:"-:all"`
}
