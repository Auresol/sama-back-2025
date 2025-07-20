package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models" // Adjust import path
)

// SchoolRepository handles database operations for the School model.
type SchoolRepository struct {
	db *gorm.DB
}

// NewSchoolRepository creates a new instance of SchoolRepository.
func NewSchoolRepository() *SchoolRepository {
	return &SchoolRepository{
		db: GetDB(), // Get the GORM DB instance from the global GetDB function
	}
}

// CreateSchool creates a new school record in the database.
func (r *SchoolRepository) CreateSchool(school *models.School) error {
	return r.db.Create(school).Error
}

// GetSchoolByID retrieves a school by its primary ID.
func (r *SchoolRepository) GetSchoolByID(id uint) (*models.School, error) {
	var school models.School
	err := r.db.First(&school, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("school with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve school by ID: %w", err)
	}
	return &school, nil
}

// GetSchoolByEmail retrieves a school by its unique email.
func (r *SchoolRepository) GetSchoolByEmail(email string) (*models.School, error) {
	var school models.School
	err := r.db.Where("email = ?", email).First(&school).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("school with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to retrieve school by email: %w", err)
	}
	return &school, nil
}

// GetSchoolByShortName retrieves a school by its unique short name.
func (r *SchoolRepository) GetSchoolByShortName(shortName string) (*models.School, error) {
	var school models.School
	err := r.db.Where("short_name = ?", shortName).First(&school).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("school with short name %s not found", shortName)
		}
		return nil, fmt.Errorf("failed to retrieve school by short name: %w", err)
	}
	return &school, nil
}

// GetAllSchools retrieves all schools with pagination.
func (r *SchoolRepository) GetAllSchools(limit, offset int) ([]models.School, error) {
	var schools []models.School
	err := r.db.Limit(limit).Offset(offset).Find(&schools).Error
	return schools, err
}

// UpdateSchool updates an existing school record.
func (r *SchoolRepository) UpdateSchool(school *models.School) error {
	// Use Save for full updates or Select/Omit for partial updates
	return r.db.Save(school).Error
}

// DeleteSchool deletes a school record by its ID.
func (r *SchoolRepository) DeleteSchool(id uint) error {
	result := r.db.Delete(&models.School{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete school: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("school with ID %d not found for deletion", id)
	}
	return nil
}

// CountSchools returns the total number of school records.
func (r *SchoolRepository) CountSchools() (int64, error) {
	var count int64
	err := r.db.Model(&models.School{}).Count(&count).Error
	return count, err
}
