package service

import (
	"context"
	"errors"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type CheckInService interface {
	CreateCheckIn(ctx context.Context, userID uint, req *CreateCheckInRequest) (*models.CheckIn, error)
	CompleteCheckIn(ctx context.Context, userID uint, id uint, req *CompleteCheckInRequest) (*models.CheckIn, error)
	UpdateCheckIn(ctx context.Context, userID uint, id uint, req *UpdateCheckInRequest) (*models.CheckIn, error)
	GetCheckIn(ctx context.Context, userID uint, id uint) (*models.CheckIn, error)
	ListCheckIns(ctx context.Context, userID uint, page, limit int, status *models.CheckInStatus) ([]models.CheckIn, int64, error)
	DeleteCheckIn(ctx context.Context, userID uint, id uint) error
	GetUserStats(ctx context.Context, userID uint) (*CheckInStats, error)
}

type CreateCheckInRequest struct {
	RouteID     uint       `json:"route_id"`
	StartStopID uint       `json:"start_stop_id"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
}

type CompleteCheckInRequest struct {
	EndStopID *uint      `json:"end_stop_id,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

type UpdateCheckInRequest struct {
	Notes *string `json:"notes,omitempty"`
}

type CheckInStats struct {
	TotalCheckIns      int     `json:"total_check_ins"`
	CompletedCheckIns  int     `json:"completed_check_ins"`
	TotalPointsEarned  int     `json:"total_points_earned"`
	CurrentStreak      int     `json:"current_streak_days"`
	LongestStreak      int     `json:"longest_streak_days"`
	UniqueRoutesCount  int     `json:"unique_routes_count"`
	TotalDurationSecs  int64   `json:"total_duration_seconds"`
	AverageDurationSecs float64 `json:"average_duration_seconds"`
}

type checkInService struct {
	checkInRepo   repository.CheckInRepository
	routeRepo     repository.RouteRepository
	stopRepo      repository.StopRepository
	userRepo      repository.UserRepository
	reputationSvc ReputationService
	badgeSvc      BadgeService
	cache         cache.Cache
}

func NewCheckInService(
	checkInRepo repository.CheckInRepository,
	routeRepo repository.RouteRepository,
	stopRepo repository.StopRepository,
	userRepo repository.UserRepository,
	reputationSvc ReputationService,
	badgeSvc BadgeService,
	cacheInstance cache.Cache,
) CheckInService {
	return &checkInService{
		checkInRepo:   checkInRepo,
		routeRepo:     routeRepo,
		stopRepo:      stopRepo,
		userRepo:      userRepo,
		reputationSvc: reputationSvc,
		badgeSvc:      badgeSvc,
		cache:         cacheInstance,
	}
}

func (s *checkInService) CreateCheckIn(ctx context.Context, userID uint, req *CreateCheckInRequest) (*models.CheckIn, error) {
	// Check for existing in-progress check-in
	existing, err := s.checkInRepo.FindInProgressByUserID(userID)
	if err == nil && existing != nil {
		return nil, ErrActiveCheckInExists
	}

	// Validate route exists
	_, err = s.routeRepo.FindByID(req.RouteID)
	if err != nil {
		return nil, ErrRouteNotFound
	}

	// Validate stop exists
	_, err = s.stopRepo.FindByID(req.StartStopID)
	if err != nil {
		return nil, ErrStopNotFound
	}

	startTime := time.Now()
	if req.StartTime != nil {
		startTime = *req.StartTime
	}

	checkIn := &models.CheckIn{
		UserID:      userID,
		RouteID:     req.RouteID,
		StartStopID: req.StartStopID,
		StartTime:   startTime,
		Notes:       req.Notes,
		Status:      models.CheckInStatusInProgress,
	}

	if err := s.checkInRepo.Create(checkIn); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateStatsCache(ctx, userID)

	return checkIn, nil
}

func (s *checkInService) CompleteCheckIn(ctx context.Context, userID uint, id uint, req *CompleteCheckInRequest) (*models.CheckIn, error) {
	checkIn, err := s.checkInRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if checkIn.UserID != userID {
		return nil, ErrNotAuthorized
	}

	// Verify status
	if checkIn.Status != models.CheckInStatusInProgress {
		return nil, ErrCheckInNotInProgress
	}

	// Validate end stop (only if provided)
	if req.EndStopID != nil {
		_, err = s.stopRepo.FindByID(*req.EndStopID)
		if err != nil {
			return nil, ErrStopNotFound
		}
	}

	endTime := time.Now()
	if req.EndTime != nil {
		endTime = *req.EndTime
	}

	// Calculate duration
	duration := int(endTime.Unix() - checkIn.StartTime.Unix())

	// Update check-in
	checkIn.EndStopID = req.EndStopID
	checkIn.EndTime = &endTime
	checkIn.Duration = &duration
	checkIn.Status = models.CheckInStatusCompleted

	// Award points (5 points per completed journey)
	points := 5
	checkIn.PointsEarned = points

	if err := s.checkInRepo.Update(checkIn); err != nil {
		return nil, err
	}

	// Add reputation points
	if err := s.reputationSvc.AddPoints(userID, points); err != nil {
		// Log but don't fail
	}

	// Check and award badges
	if err := s.badgeSvc.CheckAndAwardBadges(userID); err != nil {
		// Log but don't fail
	}

	// Invalidate cache
	s.invalidateStatsCache(ctx, userID)

	return checkIn, nil
}

func (s *checkInService) UpdateCheckIn(ctx context.Context, userID uint, id uint, req *UpdateCheckInRequest) (*models.CheckIn, error) {
	checkIn, err := s.checkInRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if checkIn.UserID != userID {
		return nil, ErrNotAuthorized
	}

	if req.Notes != nil {
		checkIn.Notes = req.Notes
	}

	if err := s.checkInRepo.Update(checkIn); err != nil {
		return nil, err
	}

	return checkIn, nil
}

func (s *checkInService) GetCheckIn(ctx context.Context, userID uint, id uint) (*models.CheckIn, error) {
	checkIn, err := s.checkInRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if checkIn.UserID != userID {
		return nil, ErrNotAuthorized
	}

	return checkIn, nil
}

func (s *checkInService) ListCheckIns(ctx context.Context, userID uint, page, limit int, status *models.CheckInStatus) ([]models.CheckIn, int64, error) {
	offset := (page - 1) * limit
	return s.checkInRepo.FindByUserID(userID, offset, limit, status)
}

func (s *checkInService) DeleteCheckIn(ctx context.Context, userID uint, id uint) error {
	checkIn, err := s.checkInRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Verify ownership
	if checkIn.UserID != userID {
		return ErrNotAuthorized
	}

	return s.checkInRepo.Delete(id)
}

func (s *checkInService) GetUserStats(ctx context.Context, userID uint) (*CheckInStats, error) {
	// Try cache first
	if s.cache != nil {
		cacheKey := "checkin_stats:" + string(rune(userID))
		var stats CheckInStats
		if err := s.cache.Get(ctx, cacheKey, &stats); err == nil {
			return &stats, nil
		}
	}

	completedCount, err := s.checkInRepo.CountCompletedByUserID(userID)
	if err != nil {
		return nil, err
	}

	uniqueRoutes, err := s.checkInRepo.CountUniqueRoutesByUserID(userID)
	if err != nil {
		return nil, err
	}

	totalDuration, err := s.checkInRepo.SumDurationByUserID(userID)
	if err != nil {
		return nil, err
	}

	consecutiveDays, err := s.checkInRepo.GetConsecutiveDaysCount(userID)
	if err != nil {
		consecutiveDays = 0
	}

	// Get user's total points earned
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	stats := &CheckInStats{
		CompletedCheckIns:  int(completedCount),
		UniqueRoutesCount:  int(uniqueRoutes),
		TotalDurationSecs:  totalDuration,
		TotalPointsEarned:  user.ReputationPoints,
		CurrentStreak:      consecutiveDays,
		LongestStreak:      consecutiveDays, // TODO: Store longest streak separately
	}

	if completedCount > 0 {
		stats.AverageDurationSecs = float64(totalDuration) / float64(completedCount)
	}

	// Cache the stats
	if s.cache != nil {
		cacheKey := "checkin_stats:" + string(rune(userID))
		s.cache.Set(ctx, cacheKey, stats, time.Hour)
	}

	return stats, nil
}

func (s *checkInService) invalidateStatsCache(ctx context.Context, userID uint) {
	if s.cache != nil {
		cacheKey := "checkin_stats:" + string(rune(userID))
		s.cache.Delete(ctx, cacheKey)
	}
}

// Error definitions
var (
	ErrActiveCheckInExists  = errors.New("active check-in already exists")
	ErrCheckInNotInProgress = errors.New("check-in is not in progress")
	ErrRouteNotFound        = errors.New("route not found")
	ErrStopNotFound         = errors.New("stop not found")
	ErrNotAuthorized        = errors.New("unauthorized: you can only access your own check-ins")
)
