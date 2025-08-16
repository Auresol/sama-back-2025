package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/pkg"
	"sama/sama-backend-2025/src/repository"
	"sama/sama-backend-2025/src/utils"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/go-playground/validator/v10"
)

// SchoolService handles business logic for schools.
type SchoolService struct {
	schoolRepo   *repository.SchoolRepository
	userRepo     *repository.UserRepository
	activityRepo *repository.ActivityRepository
	s3Client     *pkg.S3Client
	validator    *validator.Validate
}

// NewSchoolService creates a new instance of SchoolService.
func NewSchoolService(s3Client *pkg.S3Client, validate *validator.Validate) *SchoolService {
	return &SchoolService{
		schoolRepo:   repository.NewSchoolRepository(),
		userRepo:     repository.NewUserRepository(),
		activityRepo: repository.NewActivityRepository(),
		s3Client:     s3Client,
		validator:    validate,
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

	newSemesterList := models.SemesterYearList{
		strconv.Itoa(int(school.SchoolYear)) + "/" + strconv.Itoa(int(school.Semester)),
	}
	school.AvaliableSemesterList = newSemesterList

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
func (s *SchoolService) GetAllSchools(limit, offset int) ([]models.School, int, error) {
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

// // UpdateSchool updates an existing school's information.
func (s *SchoolService) GetSchoolStatisticByID(id uint, classroom string, activityIDs []uint, semester, schoolYear uint) ([]models.UserWithFinishedPercent, int, int, error) {

	// if either semester of school year is invalid, get current semester and year
	if semester == 0 || schoolYear == 0 {
		var err error
		semester, schoolYear, err = s.schoolRepo.GetSchoolSemesterAndSchoolYearByID(id)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// -1 on offset and limit to cancle pagination
	users, _, err := s.userRepo.GetUsersBySchoolID(id, 0, "", "STD", classroom, -1, -1)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to get users: %w", err)
	}

	var fisnishedAmount int

	// New array to store user with their stats and filter out who doesn't belong
	var userWithStatPos int
	usersWithStat := make([]models.UserWithFinishedPercent, len(users))

	for _, user := range users {
		// activity will sorted by it's id assending
		activities, err := s.activityRepo.GetAssignedActivitiesByUserID(user.ID, id, semester, schoolYear, false)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to retrieve statistic of user with id %d: %w", user.ID, err)
		}

		var pos int
		var sum, filterCount float32

		// since activityIDs and activity is sorted by id ascending
		// the filter algorithm apply here will be O(1)
		for _, activity := range activities {

			// Move the cursor forward until activitiyIDs[pos] is equal or greater than activity.ID
			for pos < len(activityIDs) && activityIDs[pos] < activity.ID {
				pos++
			}

			// Reach the end of filter, meaning no more activity will be apply
			if pos >= len(activityIDs) {
				break
			}

			// If the activityIDs existed in the filter, apply summation
			if activityIDs[pos] == activity.ID {
				sum += activity.FinishedPercentage
				filterCount += 1
			}
		}

		// Only apply this user if at least one activity is presented
		if filterCount > 0 {
			usersWithStat[userWithStatPos].User = user
			usersWithStat[userWithStatPos].FinishedPercent = utils.NormallizePercent(sum / filterCount)
			if usersWithStat[userWithStatPos].FinishedPercent == 100 {
				fisnishedAmount++
			}

			userWithStatPos++
		}
	}

	return usersWithStat[:userWithStatPos], fisnishedAmount, userWithStatPos - fisnishedAmount, nil
}

// GetSchoolByShortName retrieves a school by its short name.
func (s *SchoolService) GetSchoolStatisticFileByID(ctx context.Context, id uint, classroom string, activityIDs []uint, semester, schoolYear uint) (*v4.PresignedHTTPRequest, error) {

	school, err := s.schoolRepo.GetSchoolByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve school with id %d: %w", id, err)
	}

	// if either semester of school year is invalid, get current semester and year
	if semester == 0 || schoolYear == 0 {
		semester = school.Semester
		schoolYear = school.SchoolYear
	}

	filepath := school.ShortName + "_summary.xlsx"

	// TODO: generate excel file to filepath

	request, err := s.s3Client.GetPresignedDownloadURL(ctx, filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned download URL from S3 client: %w", err)
	}

	return request, nil
}

// DeleteSchool deletes a school by its ID.
func (s *SchoolService) DeleteSchool(id uint) error {
	return s.schoolRepo.DeleteSchool(id)
}

// CountSchools returns the total number of schools.
func (s *SchoolService) CountSchools() (int64, error) {
	return s.schoolRepo.CountSchools()
}
