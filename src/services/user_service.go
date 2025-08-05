package services

import (
	"fmt"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/repository"

	"github.com/go-playground/validator/v10"
)

// userService handles business logic for user accounts.
type UserService struct {
	userRepo   *repository.UserRepository
	validator  *validator.Validate
	jwtSecret  string // JWT secret for token generation
	jwtExpMins int    // JWT expiration in minutes
}

// NewuserService creates a new instance of userService.
func NewUserService(validate *validator.Validate) *UserService {
	return &UserService{
		userRepo:  repository.NewUserRepository(),
		validator: validate,
	}
}

// GetUserByID retrieves a user by ID.
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetUserByID(id)
}

// GetUserByEmail retrieves a user by email.
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

// GetAllUsers retrieves all users with pagination.
// This might be restricted to ADMIN/Sama Crew roles in the handler layer.
// func (s *UserService) GetAllUsers(limit, offset int) ([]models.User, error) {
// 	return s.userRepo.GetAllUsers(limit, offset)
// }

// GetUsersBySchoolID retrieves users for a specific school.
// This is for ADMINs to access users within their school.
func (s *UserService) GetUsersBySchoolID(schoolID, userID uint, name, role string, limit, offset int) ([]models.User, int, error) {
	return s.userRepo.GetUsersBySchoolID(schoolID, userID, name, role, limit, offset)
}

// UpdateUserProfile updates a user's profile information.
// This method handles general profile updates, not password changes.
func (s *UserService) UpdateUserProfile(user *models.User) error {
	// Crucial: Prevent password from being overwritten by an empty string
	// The password field in models.User should have `json:"-"` and `gorm:"column:password"`
	// to avoid it being marshaled/unmarshaled from JSON and to store the hashed value.
	// If you're passing a models.User struct from a request, ensure its Password field is empty.
	user.Password = ""

	// Fetch existing user to ensure we're updating a valid record
	existingUser, err := s.userRepo.GetUserByID(user.ID)
	if err != nil {
		return fmt.Errorf("user not found for update: %w", err)
	}

	// Manually update fields that are allowed to be updated from the `user` input
	// This prevents overwriting fields not intended for update or sensitive fields.
	// You might want to make this more granular based on what fields are allowed to be changed.
	existingUser.Email = user.Email // Email might be updated, but usually has unique constraint
	existingUser.Phone = user.Phone
	existingUser.Firstname = user.Firstname
	existingUser.Lastname = user.Lastname
	existingUser.ProfilePictureURL = user.ProfilePictureURL
	existingUser.Classroom = user.Classroom
	existingUser.Number = user.Number
	existingUser.Language = user.Language
	existingUser.BookmarkUserIDs = user.BookmarkUserIDs
	// Role and SchoolID might require specific permissions to change and should be handled carefully

	// Validate the updated existingUser struct before saving
	// if err := s.validator.Struct(existingUser); err != nil {
	// 	return fmt.Errorf("validation failed for updated user: %w", err)
	// }

	return s.userRepo.UpdateUser(existingUser)
}

// UpdateProfilePicture updates a user's profile picture URL.
func (s *UserService) UpdateProfilePicture(userID uint, pictureURL string) error {
	return s.userRepo.UpdateUserProfilePicture(userID, pictureURL)
}

// DeleteProfilePicture removes a user's profile picture URL.
func (s *UserService) DeleteProfilePicture(userID uint) error {
	return s.userRepo.DeleteUserProfilePicture(userID)
}

// DeleteUser deletes a user by ID.
// This method needs to include authorization logic in a real app (e.g., check if user has permission to delete this ID).
func (s *UserService) DeleteUser(id uint) error {
	return s.userRepo.DeleteUser(id)
}

// GetUserCount returns the total number of users.
func (s *UserService) GetUserCount() (int64, error) {
	return s.userRepo.CountUsers()
}

// GetUserCountBySchoolID returns the total number of users for a specific school.
func (s *UserService) GetUserCountBySchoolID(schoolID uint) (int64, error) {
	return s.userRepo.CountUsersBySchoolID(schoolID)
}
