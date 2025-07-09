package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null" binding:"required,email"`
	Password  string         `json:"-" gorm:"not null" binding:"required,min=6"` // "-" means this field won't be included in JSON
	FirstName string         `json:"first_name" gorm:"not null" binding:"required,min=2"`
	LastName  string         `json:"last_name" gorm:"not null" binding:"required,min=2"`
	Role      string         `json:"role" gorm:"default:'user'" binding:"oneof=user admin"`
	Active    bool           `json:"active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
