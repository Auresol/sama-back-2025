package models

import (
	"time"
)

type Classroom struct {
	ID uint `gorm:"primarykey"`

	SchoolID  uint   `json:"school_id,omitempty" gorm:"column:school_id;not null;uniqueIndex:idx_school_classroom,priority:1"`
	Classroom string `json:"classroom,omitempty" gorm:"column:classroom;not null;uniqueIndex:idx_school_classroom,priority:2"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (Classroom) TableName() string {
	return "classrooms"
}
