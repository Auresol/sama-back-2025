package models

import (
	"time"

	"gorm.io/gorm"
)

// Activity represents a type of activity students perform, mapped to a PostgreSQL table.
type Activity struct {
	ID uint `json:"id" gorm:"primarykey"`

	SchoolID uint   `json:"school_id" validate:"required"`
	Name     string `json:"name" validate:"required"`

	Template map[string]interface{} `json:"template" gorm:"serializer:json" validate:"required"`

	IsRequired   bool   `json:"is_required" validate:"required"`
	CoverageType string `json:"coverage_type" validate:"required,oneof=ALL JUNIOR SENIOR"`

	ExclusiveClassrooms []string `json:"exclusive_classroom" validate:"required" gorm:"-:all"`
	ExclusiveStudentIDs []uint   `json:"exclusive_student_ids" validate:"required" gorm:"-:all"`

	OwnerID uint `json:"owner_id" gorm:"index" validate:"required,gt=0"` // ID of the creator (User)

	IsActive bool       `json:"is_active" validate:"required"` // Still able to create new records
	Deadline *time.Time `json:"deadline,omitempty"`            // The date when activity is closed (nullable)

	FinishedUnit   string `json:"finished_condition" validate:"required,oneof=TIMES HOURS"`
	FinishedAmount int    `json:"finished_amount" validate:"required"`
	UpdateProtocol string `json:"update_protocol,omitempty" validate:"required,oneof=RE_EVALUATE_ALL_RECORDS IGNORE_PAST_RECORDS"`

	School                    School       `json:"-" gorm:"foreignKey:SchoolID"`
	Owner                     User         `json:"-" gorm:"foreignKey:OwnerID"`
	ExclusiveStudentObjects   []*User      `json:"-" gorm:"many2many:activity_exclusive_student_ids"`
	ExclusiveClassroomObjects []*Classroom `json:"-" gorm:"many2many:activity_exclusive_classroom"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggertype:"string"`
}

// TableName specifies the table name for the Activity model.
func (Activity) TableName() string {
	return "activities"
}

// AfterFind is a GORM callback that runs after a record is found.
// It populates the `Classrooms []string` field from the `ClassroomList` association.
func (a *Activity) AfterFind(tx *gorm.DB) (err error) {
	// Ensure ClassroomList is loaded before attempting to flatten
	// This requires preloading ClassroomList in your repository's Get methods.
	for _, obj := range a.ExclusiveClassroomObjects {
		a.ExclusiveClassrooms = append(a.ExclusiveClassrooms, obj.Classroom)
	}
	return nil
}

var ACTIVITY_COVERAGE_TYPE = []string{"ALL", "JUNIOR", "SENIOR"}

// ACTIVITY_UPDATE_PROTOCOL_ENUM defines the allowed values for the 'UpdateProtocol' field.
var ACTIVITY_UPDATE_PROTOCOL_ENUM = []string{"RE_EVALUATE_ALL_RECORDS", "IGNORE_PAST_RECORDS"}

var ACTIVITY_FINISHED_UNIT = []string{"TIMES", "HOURS"}

// Activity represents a type of activity students perform, mapped to a PostgreSQL table.
type ActivityWithStatistic struct {
	Activity
	TotalCreatedRecords  int `json:"total_created_records"`
	TotalSendedRecords   int `json:"total_sended_records"`
	TotalApprovedRecords int `json:"total_approved_records"`
	TotalRejectedRecords int `json:"total_rejected_records"`
}
