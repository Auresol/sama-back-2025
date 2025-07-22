package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models"
)

// ActivityRepository handles database operations for the Activity model.
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository creates a new instance of ActivityRepository.
func NewActivityRepository() *ActivityRepository {
	return &ActivityRepository{
		db: GetDB(), // Assumes GetDB() is correctly initialized and returns a *gorm.DB instance
	}
}

// CreateActivity creates a new activity record in the database.
// It also handles associating custom students if provided.
func (r *ActivityRepository) CreateActivity(activity *models.Activity) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the activity first
		if err := tx.Create(activity).Error; err != nil {
			return fmt.Errorf("failed to create activity: %w", err)
		}

		// Handle many-to-many relationship for CustomStudentIDs
		// GORM will automatically save associations if activity.CustomStudentIDs contains valid User models
		// and the association table is set up correctly.
		// If activity.CustomStudentIDs only contains IDs, you might need to fetch User objects first.
		if activity.CoverageType == "CUSTOM" && len(activity.CustomStudentIDs) > 0 {
			// Ensure that the CustomStudentIDs are correctly associated.
			// This typically means activity.CustomStudentIDs contains *actual* User models loaded from DB,
			// or at least User models with only their ID set.
			// GORM's Create should handle this if the IDs are valid.
			// If not, explicit association might be needed:
			// for _, student := range activity.CustomStudentIDs {
			// 	if err := tx.Model(activity).Association("CustomStudentIDs").Append(&student); err != nil {
			// 		return fmt.Errorf("failed to append custom student ID: %w", err)
			// 	}
			// }
		}
		return nil
	})
}

// GetActivityByID retrieves an activity by its ID, preloading custom student IDs.
func (r *ActivityRepository) GetActivityByID(id uint) (*models.Activity, error) {
	var activity models.Activity
	err := r.db.Preload("CustomStudentIDs").First(&activity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("activity with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve activity by ID: %w", err)
	}
	return &activity, nil
}

// GetAllActivities retrieves all activities with pagination, optionally filtering by owner ID or school ID/year/semester.
// This method can be expanded for more complex filtering.
func (r *ActivityRepository) GetAllActivities(ownerID, schoolID uint, schoolYear, semester, limit, offset int) ([]models.Activity, error) {
	var activities []models.Activity
	query := r.db.Model(&models.Activity{}).Preload("CustomStudentIDs") // Always preload students

	if ownerID != 0 {
		query = query.Where("owner_id = ?", ownerID)
	}
	if schoolID != 0 {
		// Assuming User model has SchoolID and Activity is implicitly linked to School via Owner's SchoolID
		// Or if Activity model itself has a SchoolID directly (which it doesn't in your definition)
		// For now, if schoolID is provided, we might need a join or subquery based on owner's school.
		// For simplicity, let's assume filtering by SchoolYear and Semester directly linked to Activity
		// is sufficient for school-level filtering if no direct SchoolID on Activity model.
	}
	if schoolYear != 0 {
		query = query.Where("school_year = ?", schoolYear)
	}
	if semester != 0 {
		query = query.Where("semester = ?", semester)
	}

	err := query.Limit(limit).Offset(offset).Find(&activities).Error
	return activities, err
}

// UpdateActivity updates an existing activity record.
// This includes handling updates to the CustomStudentIDs association.
func (r *ActivityRepository) UpdateActivity(activity *models.Activity) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First, update the main activity fields
		if err := tx.Save(activity).Error; err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		// Handle CustomStudentIDs association update
		// Option 1: Replace all existing associations
		if activity.CoverageType == "CUSTOM" {
			// GORM's Replace takes a slice of models. If CustomStudentIDs are just IDs,
			// you need to convert them to minimal User models first.
			var userModels []models.User
			for _, user := range activity.CustomStudentIDs { // activity.CustomStudentIDs already `[]User`
				userModels = append(userModels, models.User{ID: user.ID})
			}
			if err := tx.Model(activity).Association("CustomStudentIDs").Replace(userModels); err != nil {
				return fmt.Errorf("failed to update custom student IDs: %w", err)
			}
		} else { // CoverageType is "REQUIRE" or other, clear custom students
			if err := tx.Model(activity).Association("CustomStudentIDs").Clear(); err != nil {
				return fmt.Errorf("failed to clear custom student IDs: %w", err)
			}
		}

		return nil
	})
}

// DeleteActivity deletes an activity record by its ID.
// GORM's soft delete (DeletedAt) will be applied. Associations might need explicit handling
// if you want to clean up join table entries on hard delete, but for soft delete, they remain.
func (r *ActivityRepository) DeleteActivity(id uint) error {
	result := r.db.Delete(&models.Activity{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete activity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("activity with ID %d not found for deletion", id)
	}
	return nil
}

// CountActivities returns the total number of activity records, optionally filtered.
func (r *ActivityRepository) CountActivities(ownerID, schoolID uint, schoolYear, semester int) (int64, error) {
	var count int64
	query := r.db.Model(&models.Activity{})

	if ownerID != 0 {
		query = query.Where("owner_id = ?", ownerID)
	}
	if schoolID != 0 {
		// Similar to GetAllActivities, if Activity doesn't have SchoolID directly, this might be complex
	}
	if schoolYear != 0 {
		query = query.Where("school_year = ?", schoolYear)
	}
	if semester != 0 {
		query = query.Where("semester = ?", semester)
	}

	err := query.Count(&count).Error
	return count, err
}
