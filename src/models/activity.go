package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Activity represents a type of activity students perform, mapped to a PostgreSQL table.
type Activity struct {
	gorm.Model // Provides ID, CreatedAt, UpdatedAt, DeletedAt (for soft deletes)

	TypeID string `json:"activity_type_id,omitempty" gorm:"column:activity_type_id" validate:"required"`
	Name   string `json:"activity_name,omitempty" gorm:"column:activity_name" validate:"required"`

	// Template defines the structure of data this activity should contain (JSON).
	// Example: {"fields": [{"name": "answer", "type": "SHORT_ANS"}, {"name": "date", "type": "DATE"}]}
	Template ActivityTemplate `json:"template" gorm:"column:template;type:jsonb"`

	// Coverage defines the range of students covered by this activity.
	// "REQUIRE" for entire school, "CUSTOM" for a list of student IDs.
	CoverageType     string  `json:"coverage_type" gorm:"column:coverage_type" validate:"required,oneof=REQUIRE CUSTOM"`
	CustomStudentIDs UserIDs `json:"custom_student_ids,omitempty" gorm:"column:custom_student_ids;type:integer[]" validate:"required_if=CoverageType CUSTOM"` // Only required if CoverageType is CUSTOM

	OwnerID uint `json:"owner_id,omitempty" gorm:"column:owner_id;index" validate:"required,gt=0"` // ID of the creator (User)

	IsActive     bool       `json:"is_active" gorm:"column:is_active"`                   // Still able to create new records
	InactiveDate *time.Time `json:"inactive_date,omitempty" gorm:"column:inactive_date"` // The date when activity is closed (nullable)

	// FinishedCondition defines the condition for marking the activity as finished.
	// Example: {"type": "total_hours", "value": 10} or {"type": "total_times", "value": 5}
	FinishedCondition string `json:"finished_condition" gorm:"column:finished_condition" validate:"required,oneof=TIMES HOURS"`

	// Status of the activity, validated against ACTIVITY_STATUS_ENUM.
	Status string `json:"status,omitempty" gorm:"column:status" validate:"required,oneof=REQUIRE CUSTOM"`

	// UpdateProtocol defines how records are handled when the activity is updated.
	// Validated against ACTIVITY_UPDATE_PROTOCOL_ENUM.
	UpdateProtocol string `json:"update_protocol,omitempty" gorm:"column:update_protocol" validate:"required,oneof=RE_EVALUATE_ALL_RECORDS IGNORE_PAST_RECORDS"`

	SchoolYear int `json:"school_year" gorm:"column:school_year" validate:"required,gt=0"`
	Semester   int `json:"semester" gorm:"column:semester" validate:"required,gt=0"`
}

// ActivityTemplate is a custom type for handling map[string]interface{} as JSONB,
// specifically for the activity's template definition.
type ActivityTemplate map[string]interface{}

// Value implements the driver.Valuer interface for ActivityTemplate.
func (a ActivityTemplate) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ActivityTemplate to JSON: %w", err)
	}
	return jsonBytes, nil
}

// Scan implements the sql.Scanner interface for ActivityTemplate.
func (a *ActivityTemplate) Scan(value interface{}) error {
	if value == nil {
		*a = make(ActivityTemplate)
		return nil
	}

	var jsonBytes []byte
	switch v := value.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	default:
		return errors.New(fmt.Sprintf("unsupported type for ActivityTemplate: %T", value))
	}

	if len(jsonBytes) == 0 {
		*a = make(ActivityTemplate)
		return nil
	}

	if *a == nil {
		*a = make(ActivityTemplate)
	}

	err := json.Unmarshal(jsonBytes, a)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to ActivityTemplate: %w", err)
	}
	return nil
}

// UserIDs is a custom type for handling []uint (slice of User IDs) as a PostgreSQL INTEGER array.
// This requires importing "github.com/lib/pq".
type UserIDs []uint

// // Value implements the driver.Valuer interface for UserIDs.
// // This method is called by GORM when saving the slice to the database.
// func (u UserIDs) Value() (driver.Value, error) {
// 	// Use pq.Array to handle the conversion to PostgreSQL array format
// 	return pq.Array(u).Value()
// }

// // Scan implements the sql.Scanner interface for UserIDs.
// // This method is called by GORM when loading the slice from the database.
// func (u *UserIDs) Scan(src interface{}) error {
// 	// Use pq.Array to handle the conversion from PostgreSQL array format
// 	return pq.Array(u).Scan(src)
// }

// TableName specifies the table name for the Activity model.
func (Activity) TableName() string {
	return "activities"
}

// ACTIVITY_STATUS_ENUM defines the allowed values for the 'Status' field.
var ACTIVITY_STATUS_ENUM = []string{"REQUIRE", "CUSTOM"}

// ACTIVITY_UPDATE_PROTOCOL_ENUM defines the allowed values for the 'UpdateProtocol' field.
var ACTIVITY_UPDATE_PROTOCOL_ENUM = []string{"RE_EVALUATE_ALL_RECORDS", "IGNORE_PAST_RECORDS"}
