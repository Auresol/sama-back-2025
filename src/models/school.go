package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// School represents a school entity, mapped to a PostgreSQL table.
type School struct {
	gorm.Model // Provides ID, CreatedAt, UpdatedAt, DeletedAt

	ThaiName       string      `json:"thai_name,omitempty" gorm:"column:thai_name" validate:"required"`
	EnglishName    string      `json:"english_name,omitempty" gorm:"column:english_name" validate:"required"`
	ShortName      string      `json:"short_name,omitempty" gorm:"column:short_name;uniqueIndex" validate:"required"` // Added unique index for short_name
	Email          string      `json:"email,omitempty" gorm:"column:email;uniqueIndex" validate:"required,email"`     // Added unique index for email
	Location       string      `json:"location,omitempty" gorm:"column:location" validate:"required"`
	Phone          string      `json:"phone,omitempty" gorm:"column:phone" validate:"required,e164"` // e164 for phone number validation
	ActivityStruct ActivityMap `json:"activity_struct" gorm:"column:activity_struct;type:jsonb"`

	SchoolYear int `json:"school_year" gorm:"column:school_year" validate:"required,gt=0"` // School year must be positive
	Semester   int `json:"semester" gorm:"column:semester" validate:"required,gt=0"`       // Semester must be positive
}

// ActivityMap is a custom type for handling map[string]interface{} as JSONB.
// This allows storing flexible, semi-structured data in a JSONB column.
type ActivityMap map[string]interface{}

// Value implements the driver.Valuer interface for ActivityMap.
// This method is called by GORM when saving the map to the database.
func (a ActivityMap) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ActivityMap to JSON: %w", err)
	}
	return jsonBytes, nil
}

// Scan implements the sql.Scanner interface for ActivityMap.
// This method is called by GORM when loading the map from the database.
func (a *ActivityMap) Scan(value interface{}) error {
	if value == nil {
		*a = make(ActivityMap) // Initialize to an empty map if DB value is NULL
		return nil
	}

	var jsonBytes []byte
	switch v := value.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	default:
		return errors.New(fmt.Sprintf("unsupported type for ActivityMap: %T", value))
	}

	if len(jsonBytes) == 0 {
		*a = make(ActivityMap) // Handle empty JSON as empty map
		return nil
	}

	// Ensure the map is initialized before unmarshaling
	if *a == nil {
		*a = make(ActivityMap)
	}

	err := json.Unmarshal(jsonBytes, a)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to ActivityMap: %w", err)
	}
	return nil
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (School) TableName() string {
	return "schools"
}
