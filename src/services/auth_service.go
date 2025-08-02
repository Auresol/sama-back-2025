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
type AuthService struct {
	userRepo          *repository.UserRepository
	validator         *validator.Validate
	jwtSecret         string // JWT secret for token generation
	jwtExpMins        int    // JWT expiration in minutes
	refreshJwtSecret  string // JWT secret for token generation
	refreshJwtExpMins int    // JWT expiration in minutes
}

// NewuserService creates a new instance of userService.
func NewAuthService(
	jwtSecret string,
	jwtExpMins int,
	refreshJwtSecret string,
	refreshJwtExpMins int,
	validate *validator.Validate,
) *AuthService {
	return &AuthService{
		userRepo:          repository.NewUserRepository(),
		jwtSecret:         jwtSecret,
		jwtExpMins:        jwtExpMins,
		refreshJwtSecret:  refreshJwtSecret,
		refreshJwtExpMins: refreshJwtExpMins,
		validator:         validate,
	}
}

// RegisterUser creates a new user with hashed password.
// This method is for new user registration.
func (s *AuthService) RegisterUser(user *models.User) error {
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
	// Create the user
	return s.userRepo.CreateUser(user)
}

// Login authenticates a user and returns a JWT token if successful.
// It receives email and plain-text password directly.
func (s *AuthService) Login(email, password string) (string, string, error) {
	// Basic validation for email and password format (if not done in handler)
	// For example, if you had a LoginRequest struct passed here:
	// if err := s.validator.Struct(loginReq); err != nil { return "", fmt.Errorf("validation failed: %w", err) }

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", fmt.Errorf("failed to retrieve user for login: %w", err)
	}

	// Compare password (hashed password from DB vs. plain text password from input)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid credentials") // Passwords do not match
	}

	newToken, newRefreshToken, err := s.generateNewToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate both token: %w", err)
	}

	return newToken, newRefreshToken, nil
}

// UpdatePassword updates a user's password.
// This method should be used specifically for password changes.
func (s *AuthService) UpdatePassword(userID uint, newPassword string) error {
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

// Login authenticates a user and returns a JWT token if successful.
// It receives email and plain-text password directly.
func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {

	claims, err := utils.ValidateRefreshToken(refreshToken, s.refreshJwtSecret)
	if err != nil {
		return "", "", errors.New("Invalid or expired refresh token: " + err.Error())
	}

	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", fmt.Errorf("failed to retrieve user for refresh token: %w", err)
	}

	newToken, newRefreshToken, err := s.generateNewToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate both token: %w", err)
	}

	return newToken, newRefreshToken, nil
}

// Generate new token and refresh token from user
func (s *AuthService) generateNewToken(user *models.User) (string, string, error) {
	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.SchoolID, user.Email, user.Role, s.jwtSecret, s.jwtExpMins)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate JWT refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.ID, s.refreshJwtSecret, s.refreshJwtExpMins)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return token, refreshToken, nil
}
