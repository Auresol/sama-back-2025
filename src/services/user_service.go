package services

import (
	"errors"
	"fmt"
	"regexp"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/repository"
	"sama/sama-backend-2025/src/utils"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// userService handles business logic for user accounts.
type UserService struct {
	userRepo   *repository.UserRepository
	validator  *validator.Validate
	jwtSecret  string // JWT secret for token generation
	jwtExpMins int    // JWT expiration in minutes
}

// NewuserService creates a new instance of userService.
func NewUserService(jwtSecret string, jwtExpMins int, validate *validator.Validate) *UserService {
	return &UserService{
		userRepo:   repository.NewUserRepository(),
		validator:  validate,
		jwtSecret:  jwtSecret,
		jwtExpMins: jwtExpMins,
	}
}

// RegisterUser creates a new user with hashed password.
// This method is for new user registration.
func (s *UserService) RegisterUser(user *models.User) error {
	// Validate input user data using the service's validator instance
	if err := s.validator.StructExcept(user, "School"); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if user with this email already exists
	_, err := s.userRepo.GetUserByEmail(user.Email)
	if err == nil { // User found, so email already exists
		return errors.New("user with this email already exists")
	}
	// if !errors.Is(err, gorm.ErrRecordNotFound) { // Other database error
	// 	return fmt.Errorf("failed to check existing user: %w", err)
	// }

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword) // Store hashed password

	// Set default values if not provided (e.g., IsActive)
	// Note: ProfilePictureURL is a pointer, so check for nil
	if user.ProfilePictureURL == nil {
		defaultPictureURL := "" // Or a default placeholder URL
		user.ProfilePictureURL = &defaultPictureURL
	}
	// Default to active if not explicitly set for new users
	if user.ID == 0 { // This check ensures it's a new user creation
		user.IsActive = true
	}

	// Create the user
	return s.userRepo.CreateUser(user)
}

// Login authenticates a user and returns a JWT token if successful.
// It receives email and plain-text password directly.
func (s *UserService) Login(email, password string) (string, error) {
	// Basic validation for email and password format (if not done in handler)
	// For example, if you had a LoginRequest struct passed here:
	// if err := s.validator.Struct(loginReq); err != nil { return "", fmt.Errorf("validation failed: %w", err) }

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid credentials")
		}
		return "", fmt.Errorf("failed to retrieve user for login: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return "", errors.New("user account is deactivated")
	}

	// Compare password (hashed password from DB vs. plain text password from input)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials") // Passwords do not match
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.SchoolID, user.Email, user.Role, s.jwtSecret, s.jwtExpMins)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
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
func (s *UserService) GetAllUsers(limit, offset int) ([]models.User, error) {
	return s.userRepo.GetAllUsers(limit, offset)
}

// GetUsersBySchoolID retrieves users for a specific school.
// This is for ADMINs to access users within their school.
func (s *UserService) GetUsersBySchoolID(schoolID uint, limit, offset int) ([]models.User, error) {
	return s.userRepo.GetUsersBySchoolID(schoolID, limit, offset)
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
	existingUser.IsActive = user.IsActive
	existingUser.Classroom = user.Classroom
	existingUser.Number = user.Number
	existingUser.Status = user.Status
	existingUser.Language = user.Language
	// Role and SchoolID might require specific permissions to change and should be handled carefully

	// Validate the updated existingUser struct before saving
	if err := s.validator.Struct(existingUser); err != nil {
		return fmt.Errorf("validation failed for updated user: %w", err)
	}

	return s.userRepo.UpdateUser(existingUser)
}

// UpdatePassword updates a user's password.
// This method should be used specifically for password changes.
func (s *UserService) UpdatePassword(userID uint, newPassword string) error {
	// Password must contain only alphabet, number, or "_" only
	// Regex: ^[a-zA-Z0-9_]+$
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(newPassword) {
		return errors.New("password must contain only alphabets, numbers, or underscores")
	}
	if len(newPassword) < 8 { // Example simple validation: min length
		return errors.New("password must be at least 8 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	return s.userRepo.UpdateUserPassword(userID, string(hashedPassword))
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
