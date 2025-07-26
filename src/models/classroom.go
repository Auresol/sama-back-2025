package models

import (
	"time"
)

type Classroom struct {
	ID        uint   `gorm:"primarykey"`
	SchoolID  uint   `json:"school_id,omitempty"`
	Class     uint   `json:"class"`
	Room      uint   `json:"room"`
	Classroom string `json:"classroom"`

	School     School      `json:"school"`
	Activities []*Activity `json:"activity" gorm:"many2many:activity_exclusive_classroom"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (Classroom) TableName() string {
	return "classrooms"
}
