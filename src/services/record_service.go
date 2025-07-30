package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/repository"
)

// RecordService handles business logic for records.
type RecordService struct {
	recordRepo   *repository.RecordRepository
	schoolRepo   *repository.SchoolRepository
	userRepo     *repository.UserRepository // Assuming AccountRepository handles User model
	activityRepo *repository.ActivityRepository
	validator    *validator.Validate
}

// NewRecordService creates a new instance of RecordService.
func NewRecordService(validator *validator.Validate) *RecordService {
	return &RecordService{
		recordRepo:   repository.NewRecordRepository(),
		schoolRepo:   repository.NewSchoolRepository(),
		userRepo:     repository.NewUserRepository(),
		activityRepo: repository.NewActivityRepository(),
		validator:    validator,
	}
}

// contains is a helper for enum validation.
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// validateRecordData performs custom validation beyond struct tags, including FK checks.
func (s *RecordService) validateRecordData(record *models.Record) error {
	// Validate Status against enum
	if !contains(models.STATUS_ENUM, record.Status) {
		return fmt.Errorf("invalid Status: %s", record.Status)
	}

	// Validate StudentID
	_, err := s.userRepo.GetUserByID(record.StudentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("student with ID %d not found", record.StudentID)
		}
		return fmt.Errorf("failed to validate StudentID %d: %w", record.StudentID, err)
	}

	// Validate TeacherID
	_, err = s.userRepo.GetUserByID(*record.TeacherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("teacher with ID %d not found", record.TeacherID)
		}
		return fmt.Errorf("failed to validate TeacherID %d: %w", record.TeacherID, err)
	}

	// Validate ActivityID (assuming ActivityID in Record is uint and refers to Activity.ID)
	// If ActivityID in Record is string and refers to Activity.TypeID or Activity.Name,
	// this validation logic would need to change (e.g., s.activityRepo.GetActivityByTypeID(record.ActivityID))
	_, err = s.activityRepo.GetActivityByID(record.ActivityID) // Assuming ActivityID is uint
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("activity with ID %d not found", record.ActivityID)
		}
		return fmt.Errorf("failed to validate ActivityID %d: %w", record.ActivityID, err)
	}

	return nil
}

// CreateRecord creates a new record after validation.
func (s *RecordService) CreateRecord(record *models.Record, schoolID uint, createdByUserID uint) error {

	activity, err := s.activityRepo.GetActivityByID(record.ActivityID)
	if err != nil {
		return fmt.Errorf("failed to retrieve school with id %d: %w", schoolID, err)
	}

	if activity.SchoolID != schoolID {
		return fmt.Errorf("school id in activity and school id in your token mismatch")
	}

	school, err := s.schoolRepo.GetSchoolByID(schoolID)
	if err != nil {
		return fmt.Errorf("failed to retrieve school with id %d: %w", schoolID, err)
	}

	record.Semester = school.Semester
	record.SchoolYear = school.SchoolYear

	// Validate input using struct tags
	// if err := s.validator.Struct(record); err != nil {
	// 	return fmt.Errorf("validation failed: %w", err)
	// }

	// // Perform custom validations including FK checks
	// if err := s.validateRecordData(record); err != nil {
	// 	return fmt.Errorf("record data validation failed: %w", err)
	// }

	// Initialize StatusLogs with the initial status
	record.StatusLogs = append(record.StatusLogs, models.StatusHistory{
		Status:     record.Status,
		UpdateTime: time.Now(),
	})

	return s.recordRepo.CreateRecord(record)
}

// GetRecordByID retrieves a record by its ID.
func (s *RecordService) GetRecordByID(id uint) (*models.Record, error) {
	return s.recordRepo.GetRecordByID(id)
}

// GetAllRecords retrieves all records with filtering and pagination.
func (s *RecordService) GetAllRecords(
	studentID, teacherID, activityID uint,
	status string,
	limit, offset int,
) ([]models.Record, error) {
	return s.recordRepo.GetAllRecords(studentID, teacherID, activityID, status, limit, offset)
}

// UpdateRecord updates an existing record.
func (s *RecordService) UpdateRecord(record *models.Record, updatedByUserID uint) error {
	// Fetch existing record to ensure it exists and to get its current state for status logging
	existingRecord, err := s.recordRepo.GetRecordByID(record.ID)
	if err != nil {
		return fmt.Errorf("record not found for update: %w", err)
	}

	// Apply updates from the input `record` to `existingRecord`
	// Only update fields that are explicitly provided or allowed to be changed.

	// Handle Status change and log to history
	if record.Status != "" && existingRecord.Status != record.Status {
		existingRecord.Status = record.Status
		// Append new status to history
		newEntry := models.StatusHistory{
			Status:     record.Status,
			UpdateTime: time.Now(),
			// UserID:    &updatedByUserID, // Add UserID to StatusHistory if you want to log who updated
		}
		existingRecord.StatusLogs = append(existingRecord.StatusLogs, newEntry)
	} else if record.Status == "" {
		// If status is not provided in update, retain existing status.
		// No change, so no new history entry for status.
	}

	// Update other fields if provided in the input `record`
	// Note: SchoolID, StudentID, TeacherID, ActivityID, SchoolYear, Semester are typically
	// not updated after creation, or require specific business logic for updates.
	// For this example, I'll allow updates if provided, but you might restrict this.
	if record.ActivityID != 0 { // Assuming ActivityID is uint
		existingRecord.ActivityID = record.ActivityID
	}
	if record.Data != nil { // Check if Data map is provided
		existingRecord.Data = record.Data
	}
	if record.Advise != nil {
		existingRecord.Advise = record.Advise
	}
	if record.StudentID != 0 {
		existingRecord.StudentID = record.StudentID
	}
	// if record.TeacherID != 0 {
	// 	existingRecord.TeacherID = record.TeacherID
	// }
	if record.SchoolYear != 0 {
		existingRecord.SchoolYear = record.SchoolYear
	}
	if record.Semester != 0 {
		existingRecord.Semester = record.Semester
	}
	if record.Amount != 0 { // Assuming 0 is not a valid amount or you handle it specifically
		existingRecord.Amount = record.Amount
	}
	// StatusLogs is updated internally by service, not directly from DTO
	// existingRecord.StatusLogs = record.StatusLogs // DO NOT directly assign from DTO

	// Validate the updated existingRecord struct (including its tags)
	if err := s.validator.Struct(existingRecord); err != nil {
		return fmt.Errorf("validation failed for updated record: %w", err)
	}

	// Perform custom validations again for the updated data
	if err := s.validateRecordData(existingRecord); err != nil {
		return fmt.Errorf("updated record data validation failed: %w", err)
	}

	return s.recordRepo.UpdateRecord(existingRecord)
}

// DeleteRecord deletes a record by its ID.
func (s *RecordService) DeleteRecord(id uint) error {
	return s.recordRepo.DeleteRecord(id)
}

func (r *RecordService) SendRecord(id, teacherID, userID uint) error {
	existingRecord, err := r.recordRepo.GetRecordByID(id)
	if err != nil {
		return fmt.Errorf("record not found for update: %w", err)
	}

	existingRecord.Status = "SENDED"
	existingRecord.TeacherID = &teacherID
	existingRecord.StatusLogs = append(existingRecord.StatusLogs,
		models.StatusHistory{
			Status:     "SENDED",
			UpdateTime: time.Now(),
		})

	return r.recordRepo.UpdateRecord(existingRecord)
}

func (r *RecordService) UnsendRecord(id, userID uint) error {
	existingRecord, err := r.recordRepo.GetRecordByID(id)
	if err != nil {
		return fmt.Errorf("record not found for update: %w", err)
	}

	existingRecord.Status = "CREATED"
	existingRecord.TeacherID = nil
	existingRecord.StatusLogs = append(existingRecord.StatusLogs,
		models.StatusHistory{
			Status:     "CREATED",
			UpdateTime: time.Now(),
		})

	return r.recordRepo.UpdateRecord(existingRecord)
}

func (r *RecordService) ApproveRecord(id uint, advice *string, userID uint) error {
	existingRecord, err := r.recordRepo.GetRecordByID(id)
	if err != nil {
		return fmt.Errorf("record not found for update: %w", err)
	}

	existingRecord.Status = "APPROVED"
	existingRecord.Advise = advice
	existingRecord.StatusLogs = append(existingRecord.StatusLogs,
		models.StatusHistory{
			Status:     "APPROVED",
			UpdateTime: time.Now(),
		})

	return r.recordRepo.UpdateRecord(existingRecord)
}

func (r *RecordService) RejectRecord(id uint, advice *string, userID uint) error {
	existingRecord, err := r.recordRepo.GetRecordByID(id)
	if err != nil {
		return fmt.Errorf("record not found for update: %w", err)
	}

	existingRecord.Status = "REJECTED"
	existingRecord.Advise = advice
	existingRecord.StatusLogs = append(existingRecord.StatusLogs,
		models.StatusHistory{
			Status:     "REJECTED",
			UpdateTime: time.Now(),
		})

	return r.recordRepo.UpdateRecord(existingRecord)
}
