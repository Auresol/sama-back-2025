package models

import "time"

type OTP struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expired_at"`

	User User `json:"user"`
}

// TableName specifies the table name for the School model.
// GORM by default pluralizes struct names, but explicit naming is good practice.
func (OTP) TableName() string {
	return "otps"
}
