package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models" // Adjust import path
	"sama/sama-backend-2025/src/utils"
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

// CreateSchool creates a new school record and its associated classrooms in a transaction.
func (r *SchoolRepository) CreateSchool(school *models.School) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create the School
		if err := tx.Create(school).Error; err != nil {
			return fmt.Errorf("failed to create school: %w", err)
		}

		// 2. Create associated Classrooms
		for _, name := range school.Classrooms {

			class, room := utils.ClassroomSplit(name)

			classroom := &models.Classroom{
				SchoolID: school.ID, // Assign the ID of the newly created school
				Class:    uint(class),
				Room:     uint(room),
			}
			if err := tx.Create(classroom).Error; err != nil {
				// If a classroom fails to create (e.g., duplicate name for this school),
				// the transaction will be rolled back.
				return fmt.Errorf("failed to create classroom '%s' for school ID %d: %w", name, school.ID, err)
			}
		}

		return nil // Return nil to commit the transaction
	})
}

// GetSchoolByID retrieves a school by its primary ID.
func (r *SchoolRepository) GetSchoolByID(id uint) (*models.School, error) {
	var school models.School
	err := r.db.Preload("ClassroomList").First(&school, id).Error

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
	err := r.db.Limit(limit).Offset(offset).Preload("ClassroomList").Find(&schools).Error

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
