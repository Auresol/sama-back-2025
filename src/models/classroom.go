package models

import (
	"time"

	"gorm.io/gorm"
)

type Classroom struct {
	ID        uint   `gorm:"primarykey"`
	SchoolID  uint   `json:"school_id" gorm:"uniqueIndex:idx_classroom,priority:1" validate:"required"`
	Classroom string `json:"classroom" gorm:"uniqueIndex:idx_classroom,priority:2" validate:"required"`
	IsJunior  bool   `json:"-"`

	School     School      `json:"-"`
	Activities []*Activity `json:"-" gorm:"many2many:activity_exclusive_classroom"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggertype:"string"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (Classroom) TableName() string {
	return "classrooms"
}
