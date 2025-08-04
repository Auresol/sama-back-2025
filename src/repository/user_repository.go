package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	return r.db.Transaction(func(tx *gorm.DB) error {

		// 1. Get classroom (if exised)
		if user.Classroom != nil {
			classroom := models.Classroom{}
			if err := tx.Where("school_id = ? AND classroom = ?", user.SchoolID, user.Classroom).First(&classroom).Error; err != nil {
				return fmt.Errorf("failed to retrieve user's classroom: %w", err)
			}

			user.ClassroomID = &classroom.ID
		}

		user.BookmarkUsers = make([]models.User, len(user.BookmarkUserIDs))
		// Get student's id first
		for i, id := range user.BookmarkUserIDs {
			if err := tx.Select("id").First(&user.BookmarkUsers[i], "id = ?", id).Error; err != nil {
				return fmt.Errorf("failed to find user with id %d: %w", id, err)
			}
		}

		// 2. Create user
		if err := tx.Omit("BookmarkUsers.*").Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	})
}

// GetUserByID retrieves a user by ID.
func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Model(&models.User{}).Joins("School").Joins("ClassroomObject").First(&user, id).Error
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
	err := r.db.Model(&models.User{}).Joins("School").Joins("ClassroomObject").First(&user, "email = ?", email).Error
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
func (r *UserRepository) GetUsersBySchoolID(schoolID, userID uint, role string, limit, offset int) ([]models.User, error) {
	var users []models.User
	// Start building the query
	query := r.db.Model(&models.User{}).Joins("ClassroomObject")

	// Apply school_id filter
	query = query.Where("users.school_id = ?", schoolID)

	// Apply role filter if provided
	if role != "" {
		query = query.Where("users.role = ?", role)
	}

	// Apply role filter if provided
	if role != "" {
		query = query.Where("users.role = ?", role)
	}

	if userID != 0 {
		// Join with the user_bookmarks table (aliased as 'ub')
		// The ON clause checks if the current 'users' row's ID is present as a 'bookmark_user_id'
		// in the 'user_bookmarks' table for the 'requestingUserID'.
		query = query.Joins("LEFT JOIN user_bookmarks ub ON ub.bookmark_user_id = users.id AND ub.user_id = ?", userID)

		// Add the custom ORDER BY clause
		// CASE WHEN ub.user_id IS NOT NULL THEN 0 ELSE 1 END:
		// If ub.user_id is NOT NULL, it means there's a matching bookmark for the requestingUser, so assign 0 (comes first).
		// Otherwise (NULL), no bookmark, so assign 1 (comes second).
		query = query.Order("CASE WHEN ub.user_id IS NOT NULL THEN 0 ELSE 1 END ASC")

		// Then, add a secondary sort order (e.g., by name or ID) for consistent ordering within bookmarked/non-bookmarked groups
		query = query.Order("users.id ASC") // Or any other consistent sort
	}

	err := query.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// UpdateUser updates an existing user's general profile information.
func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		// Check if new classroom is valid
		if user.Classroom != nil {
			classroom := models.Classroom{}
			if err := tx.First(&classroom, "school_id = ? AND classroom = ?", user.SchoolID, user.Classroom).Error; err != nil {
				return fmt.Errorf("failed to retrieve user's classroom: %w", err)
			}

			user.ClassroomID = &classroom.ID
		}

		user.BookmarkUsers = make([]models.User, len(user.BookmarkUserIDs))
		// Get student's id first
		for i, id := range user.BookmarkUserIDs {
			var temp models.User
			if err := tx.Select("id").First(&temp).Error; err != nil {
				return fmt.Errorf("failed to find user with id %d: %w", id, err)
			}
			user.BookmarkUsers[i].ID = id
		}

		// 2. Create user
		if err := tx.Omit(clause.Associations).Save(user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		// Update the link to bookmark using Replace (delete all previous link, then create every new link)
		if err := tx.Model(user).Association("BookmarkUsers").Replace(user.BookmarkUsers); err != nil {
			return fmt.Errorf("failed to update bookmark user: %w", err)
		}

		return nil
	})
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
