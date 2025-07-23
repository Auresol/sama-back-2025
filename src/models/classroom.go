package models

import (
	"time"
)

type Classroom struct {
	SchoolID  uint   `json:"school_id,omitempty" gorm:"column:school_id;primarykey;autoIncrement:false"`
	Classroom string `json:"classroom,omitempty" gorm:"column:classroom;primarykey;autoIncrement:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (Classroom) TableName() string {
	return "classrooms"
}
