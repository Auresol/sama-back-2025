package repository

import (
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"sama/sama-backend-2025/src/models" // Adjust import path
)

// SchoolRepository handles database operations for the School model.
type SchoolRepository struct {
	db *gorm.DB
}

// NewSchoolRepository creates a new instance of SchoolRepository.
func NewSchoolRepository() *SchoolRepository {
	return &SchoolRepository{
		db: GetDB(),
	}
}

// CreateSchool creates a new school record and its associated classrooms in a transaction.
func (r *SchoolRepository) CreateSchool(school *models.School) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		// Create new classroom object according to input
		for _, name := range school.Classrooms {
			school.ClassroomObjects = append(school.ClassroomObjects, models.Classroom{
				SchoolID:  school.ID,
				Classroom: name,
			})
		}

		// Create both school and classroom (associate mode)
		if err := tx.Create(school).Error; err != nil {
			return fmt.Errorf("failed to create school: %w", err)
		}

		return nil
	})
}

// GetSchoolByID retrieves a school by its primary ID.
func (r *SchoolRepository) GetSchoolByID(id uint) (*models.School, error) {
	var school models.School
	err := r.db.Preload("ClassroomObjects").First(&school, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("school with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve school by ID: %w", err)
	}
	return &school, nil
}

// GetSchoolSemesterAndSchoolYearByID retrieves a school by its primary ID.
func (r *SchoolRepository) GetSchoolSemesterAndSchoolYearByID(id uint) (uint, uint, error) {
	var school models.School
	err := r.db.Select("semester", "school_year").First(&school, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, fmt.Errorf("school with ID %d not found", id)
		}
		return 0, 0, fmt.Errorf("failed to retrieve semester and school_year by school ID: %w", err)
	}
	return school.Semester, school.SchoolYear, nil
}

// GetSchoolByEmail retrieves a school by its unique email.
func (r *SchoolRepository) GetSchoolByEmail(email string) (*models.School, error) {
	var school models.School
	err := r.db.Preload("ClassroomObjects").Where("email = ?", email).First(&school).Error
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
	err := r.db.Preload("ClassroomObjects").Where("short_name = ?", shortName).First(&school).Error
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
	err := r.db.Preload("ClassroomObjects").Limit(limit).Offset(offset).Find(&schools).Error

	return schools, err
}

// UpdateSchool updates an existing school record.
func (r *SchoolRepository) UpdateSchool(school *models.School) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		// -- Classroom update --
		// Use merge sort combined algorithm
		// Associate doesn't automatically delete (or update if no id provided). Thus, explicit algorithm is needed

		// Sort input classroom string
		sort.Slice(school.Classrooms, func(i2, j2 int) bool {
			return school.Classrooms[i2] < school.Classrooms[j2]
		})

		// Find all existed classroom
		var existedClassrooms []models.Classroom
		if err := tx.Select("id", "school_id", "classroom").Where("school_id = ?", school.ID).Find(&existedClassrooms).Error; err != nil {
			return fmt.Errorf("failed to retrieve school's classroom: %w", err)
		}

		// MUST NOT USE ASSOCIATE REPLACE SINCE CLASSROOM ID IS FORIEGN KEY TO OTHER TABLE
		// Start of classroom update
		var i, j int

		// Loop until both slice run out of classroom to operate on
		for i < len(school.Classrooms) || j < len(existedClassrooms) {

			if i >= len(school.Classrooms) || (j < len(existedClassrooms) && school.Classrooms[i] > existedClassrooms[j].Classroom) {

				// Classroom existed, but no input. So we delete it
				if err := tx.Delete(&existedClassrooms[j]).Error; err != nil {
					return fmt.Errorf("failed to delete school's classroom '%s': %w", existedClassrooms[j].Classroom, err)
				}
				j++

			} else if j >= len(existedClassrooms) || (i < len(existedClassrooms) && school.Classrooms[i] < existedClassrooms[j].Classroom) {

				// Classroom doesn't existed, creating new one
				var deletedClassroom models.Classroom

				// Find soft deleted classroom
				tx.Unscoped().Where("school_id = ? AND classroom = ?", school.ID, school.Classrooms[i]).First(&deletedClassroom)

				// Restore classroom if it benn soft deleted
				if deletedClassroom.ID != 0 {
					if err := tx.Unscoped().Model(&deletedClassroom).Update("deleted_at", nil).Error; err != nil {
						return fmt.Errorf("failed to restore school's classroom '%s': %w", school.Classrooms[i], err)
					}
					i++
					continue
				}

				// Create classroom
				if err := tx.Create(&models.Classroom{
					SchoolID:  school.ID,
					Classroom: school.Classrooms[i],
				}).Error; err != nil {
					return fmt.Errorf("failed to append school's classroom '%s': %w", school.Classrooms[i], err)
				}
				i++
			} else {
				// Both contain the same classroom, do nothing
				i++
				j++
			}
		}

		// -- end of classroom update --

		if err := tx.Omit(clause.Associations).Updates(school).Error; err != nil {
			return fmt.Errorf("failed to update school: %w", err)
		}

		return nil
	})
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
