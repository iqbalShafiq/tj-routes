package repository

import (
	"tj-routes/internal/dto"
	"tj-routes/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByOAuth(provider models.OAuthProvider, oauthID string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(offset, limit int) ([]models.User, int64, error)
	Leaderboard(limit int) ([]models.User, error)
	// Analytics methods for personalized data
	GetMostFrequentRoute(userID uint) (*dto.RouteStats, error)
	GetMostFrequentStop(userID uint) (*dto.StopStats, error)
	GetTotalCheckInDuration(userID uint) (int64, error)
	GetTotalCheckInsByUserID(userID uint) (int64, error)
	GetTotalRoutesTraveledByUserID(userID uint) (int64, error)
	GetTotalUniqueRoutesByUserID(userID uint) (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByOAuth(provider models.OAuthProvider, oauthID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("oauth_provider = ? AND oauth_id = ?", provider, oauthID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) List(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *userRepository) Leaderboard(limit int) ([]models.User, error) {
	var users []models.User
	err := r.db.Order("reputation_points DESC").
		Limit(limit).
		Find(&users).Error
	return users, err
}

// GetMostFrequentRoute returns the most frequently used route by a user based on check-ins
func (r *userRepository) GetMostFrequentRoute(userID uint) (*dto.RouteStats, error) {
	var stats dto.RouteStats

	// Use raw SQL since SubQuery is not available
	err := r.db.Raw(`
		SELECT routes.id as route_id, routes.route_number, routes.name as route_name, ci.usage_count
		FROM (
			SELECT route_id, COUNT(*) as usage_count
			FROM check_ins
			WHERE user_id = ? AND status = 'completed'
			GROUP BY route_id
			ORDER BY usage_count DESC
			LIMIT 1
		) as ci
		JOIN routes ON routes.id = ci.route_id
	`, userID).Scan(&stats).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetMostFrequentStop returns the most frequently visited stop by a user based on check-ins
func (r *userRepository) GetMostFrequentStop(userID uint) (*dto.StopStats, error) {
	var stats dto.StopStats

	// Use raw SQL since SubQuery is not available
	err := r.db.Raw(`
		SELECT stops.id as stop_id, stops.name as stop_name, ci.visit_count
		FROM (
			SELECT start_stop_id as stop_id, COUNT(*) as visit_count
			FROM check_ins
			WHERE user_id = ? AND status = 'completed' AND start_stop_id IS NOT NULL
			GROUP BY start_stop_id
			ORDER BY visit_count DESC
			LIMIT 1
		) as ci
		JOIN stops ON stops.id = ci.stop_id
	`, userID).Scan(&stats).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetTotalCheckInDuration returns the total duration of all completed check-ins for a user
func (r *userRepository) GetTotalCheckInDuration(userID uint) (int64, error) {
	var total int64
	err := r.db.Model(&models.CheckIn{}).
		Where("user_id = ? AND status = ? AND duration IS NOT NULL", userID, models.CheckInStatusCompleted).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&total).Error
	return total, err
}

// GetTotalCheckInsByUserID returns the total number of check-ins for a user
func (r *userRepository) GetTotalCheckInsByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.CheckIn{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetTotalRoutesTraveledByUserID returns the total number of routes traveled (sum of all check-ins)
func (r *userRepository) GetTotalRoutesTraveledByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.CheckIn{}).
		Where("user_id = ? AND status = ?", userID, models.CheckInStatusCompleted).
		Count(&count).Error
	return count, err
}

// GetTotalUniqueRoutesByUserID returns the number of unique routes used by a user
func (r *userRepository) GetTotalUniqueRoutesByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.CheckIn{}).
		Where("user_id = ? AND status = ?", userID, models.CheckInStatusCompleted).
		Distinct("route_id").
		Count(&count).Error
	return count, err
}

