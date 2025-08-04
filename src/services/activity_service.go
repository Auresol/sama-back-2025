package services

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/repository"
	"sama/sama-backend-2025/src/utils"
)

// ActivityService handles business logic for activities.
type ActivityService struct {
	activityRepo *repository.ActivityRepository
	schoolRepo   *repository.SchoolRepository
	userRepo     *repository.UserRepository // Need user repo to validate CustomStudentIDs
	validator    *validator.Validate
}

// NewActivityService creates a new instance of ActivityService.
func NewActivityService(validate *validator.Validate) *ActivityService {
	return &ActivityService{
		activityRepo: repository.NewActivityRepository(),
		schoolRepo:   repository.NewSchoolRepository(),
		userRepo:     repository.NewUserRepository(), // Re-using UserRepository for user validation
		validator:    validate,
	}
}

// validateActivityData performs custom validation beyond struct tags.
func (s *ActivityService) validateActivityData(activity *models.Activity) error {
	// Validate CoverageType, FinishedCondition, Status, UpdateProtocol against enums
	// if !utils.Contains(models.ACTIVITY_COVERAGE_TYPE, activity.CoverageType) { // CoverageType enum is actually Status, confusing naming in model
	// 	return fmt.Errorf("invalid CoverageType: %s", activity.CoverageType)
	// }
	if !utils.Contains(models.ACTIVITY_FINISHED_UNIT, activity.FinishedUnit) {
		return fmt.Errorf("invalid FinishedCondition: %s", activity.FinishedUnit)
	}
	if !utils.Contains(models.ACTIVITY_UPDATE_PROTOCOL_ENUM, activity.UpdateProtocol) {
		return fmt.Errorf("invalid UpdateProtocol: %s", activity.UpdateProtocol)
	}

	// Validate OwnerID exists
	owner, err := s.userRepo.GetUserByID(activity.OwnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("owner_id not found")
		}
		return fmt.Errorf("failed to validate owner_id: %w", err)
	}
	// Optionally, check if owner has appropriate role (e.g., TCH, ADMIN)
	if owner.Role != "TCH" && owner.Role != "ADMIN" && owner.Role != "SAMA" {
		return errors.New("owner must be a teacher, admin, or Sama Crew member")
	}

	return nil
}

// CreateActivity creates a new activity.
func (s *ActivityService) CreateActivity(activity *models.Activity) error {
	// Validate input using struct tags
	// if err := s.validator.Struct(activity); err != nil {
	// 	return fmt.Errorf("validation failed: %w", err)
	// }

	// Perform custom validations
	// if err := s.validateActivityData(activity); err != nil {
	// 	return fmt.Errorf("activity data validation failed: %w", err)
	// }

	// if either semester of school year is invalid, get current semester and year
	if activity.Semester == 0 || activity.SchoolYear == 0 {
		semester, schoolYear, err := s.schoolRepo.GetSchoolSemesterAndSchoolYearByID(activity.SchoolID)
		if err != nil {
			return err
		}

		activity.Semester = semester
		activity.SchoolYear = schoolYear
	}

	activity.IsActive = true

	return s.activityRepo.CreateActivity(activity)
}

// GetActivityByID retrieves an activity by its ID.
func (s *ActivityService) GetActivityByID(id uint) (*models.Activity, error) {
	return s.activityRepo.GetActivityByID(id)
}

// GetAllActivities retrieves activities with filtering and pagination.
func (s *ActivityService) GetAllActivities(ownerID, schoolID, semester, schoolYear uint, limit, offset int) ([]models.Activity, error) {
	// if either semester of school year is invalid, get current semester and year
	if semester == 0 || schoolYear == 0 {
		var err error
		semester, schoolYear, err = s.schoolRepo.GetSchoolSemesterAndSchoolYearByID(schoolID)
		if err != nil {
			return nil, err
		}
	}

	return s.activityRepo.GetAllActivities(ownerID, schoolID, semester, schoolYear, limit, offset)
}

// UpdateActivity updates an existing activity.
func (s *ActivityService) UpdateActivity(activity *models.Activity) error {
	// Fetch existing activity to ensure it exists and preserve original fields not being updated.
	_, err := s.activityRepo.GetActivityByID(activity.ID)
	if err != nil {
		return fmt.Errorf("activity not found for update: %w", err)
	}

	// // Validate the updated existingActivity struct (including its tags)
	// if err := s.validator.Struct(existingActivity); err != nil {
	// 	return fmt.Errorf("validation failed for updated activity: %w", err)
	// }

	// // Perform custom validations again for the updated data
	// if err := s.validateActivityData(existingActivity); err != nil {
	// 	return fmt.Errorf("updated activity data validation failed: %w", err)
	// }

	return s.activityRepo.UpdateActivity(activity)
}

func (r *ActivityService) GetAssignedActivitiesByUserID(userID, schoolID, semester, schoolYear uint) ([]models.ActivityWithStatistic, error) {

	// if either semester of school year is invalid, get current semester and year
	if semester == 0 || schoolYear == 0 {
		var err error
		semester, schoolYear, err = r.schoolRepo.GetSchoolSemesterAndSchoolYearByID(schoolID)
		if err != nil {
			return nil, err
		}
	}

	activities, err := r.activityRepo.GetAssignedActivitiesByUserID(userID, schoolID, semester, schoolYear)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve activities: %w", err)
	}

	return activities, nil
}

// DeleteActivity deletes an activity by its ID.
func (s *ActivityService) DeleteActivity(id uint) error {
	return s.activityRepo.DeleteActivity(id)
}
