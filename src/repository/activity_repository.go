package repository

import (
	"errors"
	"fmt"
	"reflect"

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

		// TODO: use virtual table + join everything

		activity.ExclusiveClassroomObjects = make([]*models.Classroom, len(activity.ExclusiveClassrooms))
		// Get classroom's id first
		for i, name := range activity.ExclusiveClassrooms {
			if err := tx.Select("id").Where("school_id = ? AND classroom = ?", activity.SchoolID, name).First(&activity.ExclusiveClassroomObjects[i]).Error; err != nil {
				return fmt.Errorf("failed to find classroom '%s': %w", name, err)
			}
		}

		// Create activity with exclusiveClassroom association, omit the upesrt of classroom
		err := tx.Model(activity).Omit("ExclusiveClassroomObjects.*").Create(activity).Error
		if err != nil {
			return fmt.Errorf("failed to create activity: %w", err)
		}

		return nil
	})
}

// GetActivityByID retrieves an activity by its ID, preloading custom student IDs.
func (r *ActivityRepository) GetActivityByID(id uint) (*models.Activity, error) {
	var activity models.Activity
	err := r.db.Preload("ExclusiveStudentObjects").Preload("ExclusiveClassroomObjects").First(&activity, id).Error
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
func (r *ActivityRepository) GetAllActivities(ownerID, schoolID uint, limit, offset int) ([]models.Activity, error) {
	var activities []models.Activity
	query := r.db.Model(&models.Activity{}) // Always preload students

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

	err := query.Limit(limit).Offset(offset).Find(&activities).Error
	return activities, err
}

func (r *ActivityRepository) GetAssignedActivitiesByUserID(userID, schoolID uint, semester, schoolYear int) ([]models.ActivityWithStatistic, error) {
	var activities []models.ActivityWithStatistic

	// Query all activities assigned to user based on 3 condition
	// 1. activities is for junior or senior
	// 2. activitity exclusive classroom contain classroom of user
	// 3. activity exclusive student id contain user
	query := `
		SELECT 
			ac.*,
			SUM(CASE WHEN r.status = 'CREATED' THEN r.amount ELSE 0 END) AS total_created_records,
			SUM(CASE WHEN r.status = 'SENDED' THEN r.amount ELSE 0 END) AS total_sended_records,
			SUM(CASE WHEN r.status = 'APPROVED' THEN r.amount ELSE 0 END) AS total_approved_records,
			SUM(CASE WHEN r.status = 'REJECTED' THEN r.amount ELSE 0 END) AS total_rejected_records	
		FROM activities ac
		LEFT JOIN records r ON r.activity_id = ac.id
		WHERE ac.school_id = ? and
			  r.semester = ? and
			  r.school_year = ? and
		(
		-- Condition 1: Check general coverage for the user's "junior" status
			-- We'll get the user's is_junior status from their classroom
			EXISTS (
				SELECT 1
				FROM users u_main
				JOIN classrooms cl_main ON u_main.classroom_id = cl_main.id
				WHERE u_main.id = ? -- Target user ID
				AND (
					(ac.is_for_junior = TRUE AND cl_main.is_junior = TRUE) OR
					(ac.is_for_senior = TRUE AND cl_main.is_junior = FALSE)
				)
			)
			OR
			-- Condition 2: Check if activity is explicitly assigned to user's classrooms
			EXISTS (
				SELECT 1
				FROM activity_exclusive_classroom aec
				JOIN users us ON aec.classroom_id = us.classroom_id
				WHERE aec.activity_id = ac.id
				AND us.id = ? -- Target user ID
			)
			OR
			-- Condition 3: Check if activity is explicitly assigned to the user directly
			EXISTS (
				SELECT 1
				FROM activity_exclusive_student_ids aes
				WHERE aes.activity_id = ac.id
				AND aes.user_id = ? -- Target user ID
			)
		)
		GROUP BY ac.id
		ORDER BY ac.is_required DESC, ac.id ASC
	`

	if err := r.db.Raw(query, schoolID, semester, schoolYear, userID, userID, userID).Scan(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, nil
}

// UpdateActivity updates an existing activity record.
// This includes handling updates to the CustomStudentIDs association.
func (r *ActivityRepository) UpdateActivity(activity *models.Activity) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		var existedActivity models.Activity
		if err := tx.Where("id = ?", activity.ID).First(&existedActivity).Error; err != nil {
			return fmt.Errorf("failed to find existed activity: %w", err)
		}

		// Check if template got updated and new update protocol is re-evaulate
		if !reflect.DeepEqual(existedActivity.Template, activity.Template) && activity.UpdateProtocol == "RE_EVALUATE_ALL_RECORDS" {

			// find school first
			var school models.School
			if err := tx.First(&school, "id = ?", activity.SchoolID).Error; err != nil {
				return fmt.Errorf("failed to find school id %d: %w", activity.SchoolID, err)
			}

			// reset all record status to CREATED
			err := tx.Model(&models.Record{}).Where("activity_id = ? AND semester = ? AND school_year = ?", activity.ID, school.Semester, school.SchoolYear).UpdateColumn("status", "CREATED").Error
			if err != nil {
				return fmt.Errorf("failed to update records (update protocol = re-evaulate all): %w", err)
			}
		}

		activity.ExclusiveClassroomObjects = make([]*models.Classroom, len(activity.ExclusiveClassrooms))
		// Get classroom's id first
		for i, name := range activity.ExclusiveClassrooms {
			if err := tx.Select("id").Where("school_id = ? AND classroom = ?", activity.SchoolID, name).First(&activity.ExclusiveClassroomObjects[i]).Error; err != nil {
				return fmt.Errorf("failed to find classroom '%s': %w", name, err)
			}
		}

		// Update the activity fields
		if err := tx.Omit("ExclusiveClassroomObjects").Save(activity).Error; err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		// Update the link to exclusive classroom using Replace (delete all previous link, then create every new link)
		if err := tx.Model(activity).Association("ExclusiveClassroomObjects").Replace(activity.ExclusiveClassroomObjects); err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
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
