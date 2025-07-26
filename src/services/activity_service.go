package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/repository"
	"sama/sama-backend-2025/src/utils"
)

// ActivityService handles business logic for activities.
type ActivityService struct {
	activityRepo *repository.ActivityRepository
	userRepo     *repository.UserRepository // Need user repo to validate CustomStudentIDs
	validator    *validator.Validate
}

// NewActivityService creates a new instance of ActivityService.
func NewActivityService(validate *validator.Validate) *ActivityService {
	return &ActivityService{
		activityRepo: repository.NewActivityRepository(),
		userRepo:     repository.NewUserRepository(), // Re-using UserRepository for user validation
		validator:    validate,
	}
}

// validateActivityData performs custom validation beyond struct tags.
func (s *ActivityService) validateActivityData(activity *models.Activity) error {
	// Validate CoverageType, FinishedCondition, Status, UpdateProtocol against enums
	if !utils.Contains(models.ACTIVITY_COVERAGE_TYPE, activity.CoverageType) { // CoverageType enum is actually Status, confusing naming in model
		return fmt.Errorf("invalid CoverageType: %s", activity.CoverageType)
	}
	if !utils.Contains(models.ACTIVITY_FINISHED_UNIT, activity.FinishedUnit) {
		return fmt.Errorf("invalid FinishedCondition: %s", activity.FinishedUnit)
	}
	if !utils.Contains(models.ACTIVITY_UPDATE_PROTOCOL_ENUM, activity.UpdateProtocol) {
		return fmt.Errorf("invalid UpdateProtocol: %s", activity.UpdateProtocol)
	}

	// // Conditional validation for CustomStudentIDs
	// if activity.CoverageType == "CUSTOM" {
	// 	if len(activity.CustomStudentIDs) == 0 {
	// 		return errors.New("CustomStudentIDs are required when CoverageType is CUSTOM")
	// 	}
	// 	// Validate if all CustomStudentIDs refer to existing users
	// 	for _, student := range activity.CustomStudentIDs {
	// 		_, err := s.userRepo.GetUserByID(student.ID)
	// 		if err != nil {
	// 			if errors.Is(err, gorm.ErrRecordNotFound) {
	// 				return fmt.Errorf("custom student ID %d not found", student.ID)
	// 			}
	// 			return fmt.Errorf("failed to validate custom student ID %d: %w", student.ID, err)
	// 		}
	// 	}
	// } else if activity.CoverageType == "REQUIRE" {
	// 	if len(activity.CustomStudentIDs) > 0 {
	// 		return errors.New("CustomStudentIDs must be empty when CoverageType is REQUIRE")
	// 	}
	// }

	// Validate OwnerID exists
	owner, err := s.userRepo.GetUserByID(activity.OwnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("owner_id not found")
		}
		return fmt.Errorf("failed to validate owner_id: %w", err)
	}
	// Optionally, check if owner has appropriate role (e.g., TCH, ADMIN)
	if owner.Role != "TCH" && owner.Role != "ADMIN" && owner.Role != "SAMA_CREW" {
		return errors.New("owner must be a teacher, admin, or Sama Crew member")
	}

	return nil
}

// CreateActivity creates a new activity.
func (s *ActivityService) CreateActivity(activity *models.Activity) error {
	// Validate input using struct tags
	if err := s.validator.Struct(activity); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Perform custom validations
	if err := s.validateActivityData(activity); err != nil {
		return fmt.Errorf("activity data validation failed: %w", err)
	}

	// Set default IsActive to true if not set
	if activity.ID == 0 { // For new creation
		activity.IsActive = true
	}

	return s.activityRepo.CreateActivity(activity)
}

// GetActivityByID retrieves an activity by its ID.
func (s *ActivityService) GetActivityByID(id uint) (*models.Activity, error) {
	return s.activityRepo.GetActivityByID(id)
}

// GetAllActivities retrieves activities with filtering and pagination.
func (s *ActivityService) GetAllActivities(ownerID, schoolID uint, schoolYear, semester, limit, offset int) ([]models.Activity, error) {
	return s.activityRepo.GetAllActivities(ownerID, schoolID, schoolYear, semester, limit, offset)
}

// UpdateActivity updates an existing activity.
func (s *ActivityService) UpdateActivity(activity *models.Activity) error {
	// Fetch existing activity to ensure it exists and preserve original fields not being updated.
	existingActivity, err := s.activityRepo.GetActivityByID(activity.ID)
	if err != nil {
		return fmt.Errorf("activity not found for update: %w", err)
	}

	// Apply updates from the input 'activity' to the 'existingActivity'.
	// Only update fields that are explicitly provided or allowed to be changed.
	// Note: GORM's Save method will update all fields, so careful assignment is needed.
	// Better to update specific fields or use Select() in repo if partial update is preferred.

	// For simplicity, I'll copy fields from the input `activity` to `existingActivity`
	// then validate and save the `existingActivity`.
	existingActivity.Name = activity.Name
	existingActivity.Template = activity.Template
	existingActivity.CoverageType = activity.CoverageType
	// existingActivity.CustomStudentIDs = activity.CustomStudentIDs // This will be handled by repo association
	existingActivity.IsActive = activity.IsActive
	existingActivity.FinishedUnit = activity.FinishedUnit
	existingActivity.FinishedAmount = activity.FinishedAmount
	existingActivity.UpdateProtocol = activity.UpdateProtocol

	// Handle InactiveDate logic
	if !existingActivity.IsActive && existingActivity.Deadline == nil {
		now := time.Now()
		existingActivity.Deadline = &now
	} else if existingActivity.IsActive && existingActivity.Deadline != nil {
		existingActivity.Deadline = nil // Clear InactiveDate if re-activated
	}

	// Validate the updated existingActivity struct (including its tags)
	if err := s.validator.Struct(existingActivity); err != nil {
		return fmt.Errorf("validation failed for updated activity: %w", err)
	}

	// Perform custom validations again for the updated data
	if err := s.validateActivityData(existingActivity); err != nil {
		return fmt.Errorf("updated activity data validation failed: %w", err)
	}

	return s.activityRepo.UpdateActivity(existingActivity)
}

// DeleteActivity deletes an activity by its ID.
func (s *ActivityService) DeleteActivity(id uint) error {
	return s.activityRepo.DeleteActivity(id)
}
