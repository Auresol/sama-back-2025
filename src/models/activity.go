package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Activity represents a type of activity students perform, mapped to a PostgreSQL table.
type Activity struct {
	ID uint `gorm:"primarykey"`

	Name string `json:"activity_name,omitempty" gorm:"column:activity_name" validate:"required"`

	// Template defines the structure of data this activity should contain (JSON).
	// Example: {"fields": [{"name": "answer", "type": "SHORT_ANS"}, {"name": "date", "type": "DATE"}]}
	Template ActivityTemplate `json:"template" gorm:"column:template;type:jsonb"`

	IsRequired   bool   `json:"is_required,omitempty" gorm:"column:is_required" validate:"required"`
	CoverageType string `json:"coverage_type,omitempty" gorm:"column:coverage_type" validate:"required,oneof=ALL JUNIOR SENIOR"`

	SchoolID           uint         `json:"school_id,omitempty" gorm:"column:school_id"`
	ExclusiveClassroom []*Classroom `json:"exclusive_classroom,omitempty" gorm:"column:exclusive_classroom;many2many:activity_exclusive_classroom"`

	ExclusiveStudentIDs []*User `json:"exclusive_student_ids,omitempty" gorm:"column:exclusive_student_ids;many2many:activity_exclusive_student_ids"`

	OwnerID uint `json:"owner_id,omitempty" gorm:"column:owner_id;index" validate:"required,gt=0"` // ID of the creator (User)

	IsActive     bool       `json:"is_active" gorm:"column:is_active"`                   // Still able to create new records
	InactiveDate *time.Time `json:"inactive_date,omitempty" gorm:"column:inactive_date"` // The date when activity is closed (nullable)

	FinishedUnit   string `json:"finished_condition" gorm:"column:finished_condition" validate:"required,oneof=TIMES HOURS"`
	FinishedAmount int

	// UpdateProtocol defines how records are handled when the activity is updated.
	// Validated against ACTIVITY_UPDATE_PROTOCOL_ENUM.
	UpdateProtocol string `json:"update_protocol,omitempty" gorm:"column:update_protocol" validate:"required,oneof=RE_EVALUATE_ALL_RECORDS IGNORE_PAST_RECORDS"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
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

// TableName specifies the table name for the Activity model.
func (Activity) TableName() string {
	return "activities"
}

// ACTIVITY_STATUS_ENUM defines the allowed values for the 'Status' field.
var ACTIVITY_STATUS_ENUM = []string{"REQUIRE", "CUSTOM"}

// ACTIVITY_UPDATE_PROTOCOL_ENUM defines the allowed values for the 'UpdateProtocol' field.
var ACTIVITY_UPDATE_PROTOCOL_ENUM = []string{"RE_EVALUATE_ALL_RECORDS", "IGNORE_PAST_RECORDS"}

var ACTIVITY_FINISHED_CONDITION = []string{"TIMES", "HOURS"}
