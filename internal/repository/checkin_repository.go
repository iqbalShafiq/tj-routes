package repository

import (
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type CheckInRepository interface {
	Create(checkIn *models.CheckIn) error
	FindByID(id uint) (*models.CheckIn, error)
	FindByUserID(userID uint, offset, limit int, status *models.CheckInStatus) ([]models.CheckIn, int64, error)
	FindInProgressByUserID(userID uint) (*models.CheckIn, error)
	Update(checkIn *models.CheckIn) error
	Delete(id uint) error
	CountByUserID(userID uint, status *models.CheckInStatus) (int64, error)
	CountCompletedByUserID(userID uint) (int64, error)
	CountUniqueRoutesByUserID(userID uint) (int64, error)
	GetConsecutiveDaysCount(userID uint) (int, error)
	SumDurationByUserID(userID uint) (int64, error)
}

type checkInRepository struct {
	db *gorm.DB
}

func NewCheckInRepository(db *gorm.DB) CheckInRepository {
	return &checkInRepository{db: db}
}

func (r *checkInRepository) Create(checkIn *models.CheckIn) error {
	// Omit EndStopID when nil/0 to insert NULL instead of violating FK constraint
	if checkIn.EndStopID == nil || *checkIn.EndStopID == 0 {
		return r.db.Omit("EndStopID").Create(checkIn).Error
	}
	return r.db.Create(checkIn).Error
}

func (r *checkInRepository) FindByID(id uint) (*models.CheckIn, error) {
	var checkIn models.CheckIn
	err := r.db.Preload("Route").
		Preload("StartStop").
		Preload("EndStop").
		First(&checkIn, id).Error
	if err != nil {
		return nil, err
	}
	return &checkIn, nil
}

func (r *checkInRepository) FindByUserID(userID uint, offset, limit int, status *models.CheckInStatus) ([]models.CheckIn, int64, error) {
	var checkIns []models.CheckIn
	var total int64

	query := r.db.Model(&models.CheckIn{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).
		Preload("Route").
		Preload("StartStop").
		Preload("EndStop").
		Order("start_time DESC").
		Find(&checkIns).Error

	return checkIns, total, err
}

func (r *checkInRepository) FindInProgressByUserID(userID uint) (*models.CheckIn, error) {
	var checkIn models.CheckIn
	err := r.db.Where("user_id = ? AND status = ?", userID, models.CheckInStatusInProgress).
		First(&checkIn).Error
	if err != nil {
		return nil, err
	}
	return &checkIn, nil
}

func (r *checkInRepository) Update(checkIn *models.CheckIn) error {
	return r.db.Save(checkIn).Error
}

func (r *checkInRepository) Delete(id uint) error {
	return r.db.Delete(&models.CheckIn{}, id).Error
}

func (r *checkInRepository) CountByUserID(userID uint, status *models.CheckInStatus) (int64, error) {
	var count int64
	query := r.db.Model(&models.CheckIn{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *checkInRepository) CountCompletedByUserID(userID uint) (int64, error) {
	return r.CountByUserID(userID, ptr(models.CheckInStatusCompleted))
}

func (r *checkInRepository) CountUniqueRoutesByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.CheckIn{}).
		Where("user_id = ? AND status = ?", userID, models.CheckInStatusCompleted).
		Distinct("route_id").
		Count(&count).Error
	return count, err
}

func (r *checkInRepository) GetConsecutiveDaysCount(userID uint) (int, error) {
	// Query to count consecutive days with check-ins
	// Using raw SQL for the complex date calculation
	var count int
	err := r.db.Raw(`
		WITH dates AS (
			SELECT DISTINCT DATE(start_time) as check_date
			FROM check_ins
			WHERE user_id = ? AND status = 'completed'
			ORDER BY check_date DESC
		),
		consecutive AS (
			SELECT check_date,
				   DATE_PART('day', check_date - LAG(check_date) OVER (ORDER BY check_date)) as diff
			FROM dates
		)
		SELECT COUNT(*) FROM consecutive WHERE diff = 1 OR diff IS NULL
	`, userID).Scan(&count).Error

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *checkInRepository) SumDurationByUserID(userID uint) (int64, error) {
	var sum int64
	err := r.db.Model(&models.CheckIn{}).
		Where("user_id = ? AND status = ?", userID, models.CheckInStatusCompleted).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&sum).Error
	return sum, err
}

// Helper function
func ptr[T any](v T) *T {
	return &v
}
