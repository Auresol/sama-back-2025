package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models"
)

// RecordRepository handles database operations for the Record model.
type RecordRepository struct {
	db *gorm.DB
}

// NewRecordRepository creates a new instance of RecordRepository.
func NewRecordRepository() *RecordRepository {
	return &RecordRepository{
		db: GetDB(), // Get the GORM DB instance
	}
}

// CreateRecord creates a new record in the database.
func (r *RecordRepository) CreateRecord(record *models.Record) error {
	return r.db.Create(record).Error
}

// GetRecordByID retrieves a record by its primary ID.
func (r *RecordRepository) GetRecordByID(id uint) (*models.Record, error) {
	var record models.Record
	// Preload any associations if needed (e.g., Activity, School, Student, Teacher)
	// For example: .Preload("Activity").Preload("School")...
	err := r.db.First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("record with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve record by ID: %w", err)
	}
	return &record, nil
}

// GetAllRecords retrieves all records with pagination and optional filtering.
// Filters can be added based on SchoolID, StudentID, TeacherID, ActivityID, Status etc.
func (r *RecordRepository) GetAllRecords(
	studentID, teacherID, activityID uint,
	status string,
	semester, schoolYear int,
	limit, offset int,
) ([]models.Record, int, error) {
	var records []models.Record
	var count int64
	query := r.db.Model(&models.Record{}).Where("semester = ? AND school_year = ?", semester, schoolYear)

	if studentID != 0 {
		query = query.Where("student_id = ?", studentID)
	}
	if teacherID != 0 {
		query = query.Where("teacher_id = ?", teacherID)
	}
	if activityID != 0 {
		query = query.Where("activity_id = ?", activityID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Add preloads if you want to fetch related data with the records
	// query = query.Preload("Activity").Preload("School").Preload("Student").Preload("Teacher")

	countQuery := query
	err := countQuery.Count(&count).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count records: %w", err)
	}

	err = query.Limit(limit).Offset(offset).Find(&records).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrive records: %w", err)
	}

	return records, int(count), nil
}

// UpdateRecord updates an existing record.
// This method is designed to update the entire record object, including JSONB fields.
// The service layer will handle appending to StatusLogs before calling this.
func (r *RecordRepository) UpdateRecord(record *models.Record) error {
	// Use Save to update all fields, including JSONB fields like Data and StatusLogs.
	// GORM will handle the marshaling/unmarshaling due to Value/Scan methods.
	return r.db.Save(record).Error
}

// DeleteRecord deletes a record by its ID.
func (r *RecordRepository) DeleteRecord(id uint) error {
	result := r.db.Delete(&models.Record{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record with ID %d not found for deletion", id)
	}
	return nil
}

// CountRecords returns the total number of record records, optionally filtered.
func (r *RecordRepository) CountRecords(
	studentID, teacherID, activityID uint,
	status string,
) (int, error) {
	var count int64
	query := r.db.Model(&models.Record{})

	if studentID != 0 {
		query = query.Where("student_id = ?", studentID)
	}
	if teacherID != 0 {
		query = query.Where("teacher_id = ?", teacherID)
	}
	if activityID != 0 {
		query = query.Where("activity_id = ?", activityID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	return int(count), err
}
