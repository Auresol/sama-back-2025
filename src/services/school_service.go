package services

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/repository"

	"github.com/go-playground/validator/v10"
)

// SchoolService handles business logic for schools.
type SchoolService struct {
	schoolRepo *repository.SchoolRepository
	validator  *validator.Validate
}

// NewSchoolService creates a new instance of SchoolService.
func NewSchoolService(validate *validator.Validate) *SchoolService {
	return &SchoolService{
		schoolRepo: repository.NewSchoolRepository(),
		validator:  validate,
	}
}

// CreateSchool creates a new school after validation and uniqueness checks.
func (s *SchoolService) CreateSchool(school *models.School) error {
	// Validate input school data
	if err := s.validator.Struct(school); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if a school with this email already exists
	_, err := s.schoolRepo.GetSchoolByEmail(*school.Email)
	if err == nil {
		return errors.New("school with this email already exists")
	}
	// if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 	return fmt.Errorf("failed to check existing school by email: %w", err)
	// }

	// Check if a school with this short name already exists
	_, err = s.schoolRepo.GetSchoolByShortName(school.ShortName)
	if err == nil {
		return errors.New("school with this short name already exists")
	}
	// if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 	return fmt.Errorf("failed to check existing school by short name: %w", err)
	// }

	// Create the school
	return s.schoolRepo.CreateSchool(school)
}

// GetSchoolByID retrieves a school by its ID.
func (s *SchoolService) GetSchoolByID(id uint) (*models.School, error) {
	return s.schoolRepo.GetSchoolByID(id)
}

// GetSchoolByEmail retrieves a school by its email.
func (s *SchoolService) GetSchoolByEmail(email string) (*models.School, error) {
	return s.schoolRepo.GetSchoolByEmail(email)
}

// GetSchoolByShortName retrieves a school by its short name.
func (s *SchoolService) GetSchoolByShortName(shortName string) (*models.School, error) {
	return s.schoolRepo.GetSchoolByShortName(shortName)
}

// GetAllSchools retrieves all schools with pagination.
func (s *SchoolService) GetAllSchools(limit, offset int) ([]models.School, error) {
	return s.schoolRepo.GetAllSchools(limit, offset)
}

// UpdateSchool updates an existing school's information.
func (s *SchoolService) UpdateSchool(school *models.School) error {
	// Fetch existing school to ensure it exists and to avoid overwriting unintended fields
	existingSchool, err := s.schoolRepo.GetSchoolByID(school.ID)
	if err != nil {
		return fmt.Errorf("school not found for update: %w", err)
	}

	// Apply updates from the input 'school' to the 'existingSchool'
	// Only update fields that are explicitly provided or allowed to be changed.
	existingSchool.ThaiName = school.ThaiName
	existingSchool.EnglishName = school.EnglishName
	existingSchool.ShortName = school.ShortName
	existingSchool.DefaultActivityDeadline = school.DefaultActivityDeadline
	existingSchool.Email = school.Email
	existingSchool.Location = school.Location
	existingSchool.Phone = school.Phone
	existingSchool.Classrooms = school.Classrooms

	// Validate the updated existingSchool struct before saving
	if err := s.validator.Struct(existingSchool); err != nil {
		return fmt.Errorf("validation failed for updated school: %w", err)
	}

	// Check for uniqueness if email or short name is changed
	if existingSchool.Email != school.Email {
		_, err = s.schoolRepo.GetSchoolByEmail(*school.Email)
		if err == nil {
			return errors.New("new email already exists for another school")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check new email uniqueness: %w", err)
		}
	}
	if existingSchool.ShortName != school.ShortName {
		_, err = s.schoolRepo.GetSchoolByShortName(school.ShortName)
		if err == nil {
			return errors.New("new short name already exists for another school")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check new short name uniqueness: %w", err)
		}
	}

	return s.schoolRepo.UpdateSchool(school)
}

// DeleteSchool deletes a school by its ID.
func (s *SchoolService) DeleteSchool(id uint) error {
	return s.schoolRepo.DeleteSchool(id)
}

// CountSchools returns the total number of schools.
func (s *SchoolService) CountSchools() (int64, error) {
	return s.schoolRepo.CountSchools()
}
