package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models" // Assuming this is your module path to models
)

// userRepository handles database operations for user accounts.
type UserRepository struct {
	db *gorm.DB
}

// NewuserRepository creates a new instance of userRepository.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: GetDB(), // Get the GORM DB instance
	}
}

// CreateUser creates a new user account.
// This method is used for registration by Sama Crew or ADMIN.
func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByID retrieves a user by ID.
func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve user by ID: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email.
// Useful for login authentication.
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to retrieve user by email: %w", err)
	}
	return &user, nil
}

// GetUsersBySchoolID retrieves all users belonging to a specific school with pagination.
// This supports the "only able to access data from their school" feature.
func (r *UserRepository) GetUsersBySchoolID(schoolID uint, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("school_id = ?", schoolID).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// GetAllUsers retrieves all users with pagination (potentially for Sama Crew/Global ADMIN).
func (r *UserRepository) GetAllUsers(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// UpdateUser updates an existing user's general profile information.
func (r *UserRepository) UpdateUser(user *models.User) error {
	// Use Save for full updates or Select/Omit for partial updates
	return r.db.Save(user).Error
}

// UpdateUserPassword updates a user's password.
// The password should be hashed *before* being passed to this method.
func (r *UserRepository) UpdateUserPassword(userID uint, hashedPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// UpdateUserProfilePicture updates a user's profile picture URL.
func (r *UserRepository) UpdateUserProfilePicture(userID uint, pictureURL string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("profile_picture_url", pictureURL).Error
}

// DeleteUserProfilePicture removes a user's profile picture URL.
func (r *UserRepository) DeleteUserProfilePicture(userID uint) error {
	// Set the profile_picture_url to NULL
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("profile_picture_url", gorm.Expr("NULL")).Error
}

// DeleteUser deletes a user by ID.
// This supports deletion by self, ADMIN, or Sama Crew.
func (r *UserRepository) DeleteUser(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// CountUsers returns the total number of users.
func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountUsersBySchoolID returns the total number of users for a specific school.
func (r *UserRepository) CountUsersBySchoolID(schoolID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("school_id = ?", schoolID).Count(&count).Error
	return count, err
}

// CheckEmailExistsInSchool checks if an email is already registered to a user within a specific school.
// This can be used for the "STD email must be known to school first" rule,
// though a more explicit "whitelist" table might be needed for strict pre-registration.
func (r *UserRepository) CheckEmailExistsInSchool(email string, schoolID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).
		Where("email = ? AND school_id = ?", email, schoolID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check email existence in school: %w", err)
	}
	return count > 0, nil
}

// UpdateClassroomForSchool updates the classroom for all students in a given school.
// This method is for ADMINs.
func (r *UserRepository) UpdateClassroomForSchool(schoolID uint, oldClassroom, newClassroom string) (int64, error) {
	result := r.db.Model(&models.User{}).
		Where("school_id = ? AND classroom = ?", schoolID, oldClassroom).
		Update("classroom", newClassroom)
	return result.RowsAffected, result.Error
}

// UpdateClassroomForStudent updates the classroom for a specific student.
// This might be used by ADMIN or TCH for individual student updates.
func (r *UserRepository) UpdateClassroomForStudent(studentID uint, newClassroom string) error {
	return r.db.Model(&models.User{}).Where("id = ?", studentID).Update("classroom", newClassroom).Error
}

// Add a field to the User model for ProfilePictureURL
// This should be added to your models/user.go file:
/*
type User struct {
    gorm.Model
    // ... existing fields ...
    ProfilePictureURL sql.NullString `json:"profile_picture_url,omitempty" gorm:"column:profile_picture_url"` // Use sql.NullString for nullable URL
    // ... rest of your User model
}
*/
