package repository

import (
	"fmt"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/utils"
	"time"

	"gorm.io/gorm"
)

// OTPRepository handles database operations for the OTP model.
type OTPRepository struct {
	db *gorm.DB
}

// NewOTPRepository creates a new OTP repository.
func NewOTPRepository() *OTPRepository {
	return &OTPRepository{
		db: GetDB(),
	}
}

// CreateOrUpdateOTP generates a new OTP and saves it to the database.
// It will also delete any existing OTP for the user to prevent conflicts.
func (r *OTPRepository) CreateOTP(userID uint) (*models.OTP, error) {
	// Step 1: Generate a new OTP code and set its expiration
	otpCode := utils.GenerateOTPCode()
	expiresAt := time.Now().Add(5 * time.Minute)

	// Step 2: Delete any existing OTP for the user to ensure uniqueness
	if err := r.db.Delete(&models.OTP{}, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to delete existing OTP: %w", err)
	}

	// Step 3: Create the new OTP
	otp := &models.OTP{
		UserID:    userID,
		Code:      otpCode,
		ExpiresAt: expiresAt,
	}

	if err := r.db.Create(otp).Error; err != nil {
		return nil, fmt.Errorf("failed to create new OTP: %w", err)
	}

	return otp, nil
}

// VerifyOTP checks if a given OTP code is valid and not expired.
func (r *OTPRepository) VerifyOTP(userID uint, code string) (bool, error) {
	var otp models.OTP
	result := r.db.Where("user_id = ? AND code = ?", userID, code).First(&otp)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil // Code not found
		}
		return false, fmt.Errorf("failed to query OTP: %w", result.Error)
	}

	// Check if the OTP is expired
	if time.Now().After(otp.ExpiresAt) {
		return false, nil // OTP is expired
	}

	return true, nil
}

// DeleteOTP
func (r *OTPRepository) DeleteOTP(userID uint) error {

	result := r.db.Delete(&models.OTP{}, "user_id = ?", userID)

	if result.Error != nil {
		return fmt.Errorf("failed to delete OTP: %w", result.Error)
	}

	return nil
}
