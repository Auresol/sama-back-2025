package models

import (
	"time"
)

type Classroom struct {
	ID        uint   `gorm:"primarykey" validate:"required"`
	SchoolID  uint   `json:"school_id" validate:"required"`
	Class     uint   `json:"class" validate:"required"`
	Room      uint   `json:"room" validate:"required"`
	Classroom string `json:"classroom" validate:"required"`

	School     School      `json:"-"`
	Activities []*Activity `json:"-" gorm:"many2many:activity_exclusive_classroom"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (Classroom) TableName() string {
	return "classrooms"
}
