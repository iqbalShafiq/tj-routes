package service

import (
	"context"
	"errors"
	"time"

	"tj-routes/internal/dto"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

var (
	ErrFavoriteAlreadyExists   = errors.New("favorite already exists")
	ErrFavoriteNotFound        = errors.New("favorite not found")
	ErrPlaceNotFound           = errors.New("place not found")
	ErrSavedNavigationNotFound = errors.New("saved navigation not found")
	ErrInvalidNavigationPoint  = errors.New("invalid navigation point")
	ErrUnauthorizedPlaceAccess = errors.New("unauthorized: you can only access your own places")
	ErrUnauthorizedAccess      = errors.New("unauthorized: access denied")
)

// UserPersonalizedService interface for user personalized data operations
type UserPersonalizedService interface {
	// Favorites
	AddFavoriteRoute(ctx context.Context, userID, routeID uint) error
	AddFavoriteStop(ctx context.Context, userID, stopID uint) error
	RemoveFavoriteRoute(ctx context.Context, userID, routeID uint) error
	RemoveFavoriteStop(ctx context.Context, userID, stopID uint) error
	GetFavoriteRoutes(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteRouteListResponse, error)
	GetFavoriteStops(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteStopListResponse, error)
	IsFavoriteRoute(ctx context.Context, userID, routeID uint) (bool, error)
	IsFavoriteStop(ctx context.Context, userID, stopID uint) (bool, error)

	// Places
	CreatePlace(ctx context.Context, userID uint, req *dto.CreateUserPlaceRequest) (*dto.UserPlaceResponse, error)
	UpdatePlace(ctx context.Context, userID, placeID uint, req *dto.UpdateUserPlaceRequest) (*dto.UserPlaceResponse, error)
	DeletePlace(ctx context.Context, userID, placeID uint) error
	GetPlaces(ctx context.Context, userID uint) (*dto.UserPlaceListResponse, error)
	GetPlaceByID(ctx context.Context, userID, placeID uint) (*dto.UserPlaceResponse, error)

	// Recent Views
	RecordRecentView(ctx context.Context, userID uint, req *dto.RecordRecentViewRequest) error
	GetRecentRoutes(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteRouteListResponse, error)
	GetRecentStops(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteStopListResponse, error)
	GetRecentNavigations(ctx context.Context, userID uint, page, limit int) (*dto.UserSavedNavigationListResponse, error)

	// Saved Navigations
	CreateSavedNavigation(ctx context.Context, userID uint, req *dto.CreateUserSavedNavigationRequest) (*dto.UserSavedNavigationResponse, error)
	UpdateSavedNavigation(ctx context.Context, userID, navID uint, req *dto.UpdateUserSavedNavigationRequest) (*dto.UserSavedNavigationResponse, error)
	DeleteSavedNavigation(ctx context.Context, userID, navID uint) error
	GetSavedNavigations(ctx context.Context, userID uint, page, limit int) (*dto.UserSavedNavigationListResponse, error)

	// Analytics
	GetUserAnalytics(ctx context.Context, userID uint) (*dto.UserAnalyticsResponse, error)
}

type userPersonalizedService struct {
	userFavoriteRepo      repository.UserFavoriteRepository
	userPlaceRepo         repository.UserPlaceRepository
	userRecentViewRepo    repository.UserRecentViewRepository
	userSavedNavRepo      repository.UserSavedNavigationRepository
	userRepo              repository.UserRepository
	routeRepo             repository.RouteRepository
	stopRepo              repository.StopRepository
}

// NewUserPersonalizedService creates a new UserPersonalizedService
func NewUserPersonalizedService(
	userFavoriteRepo repository.UserFavoriteRepository,
	userPlaceRepo repository.UserPlaceRepository,
	userRecentViewRepo repository.UserRecentViewRepository,
	userSavedNavRepo repository.UserSavedNavigationRepository,
	userRepo repository.UserRepository,
	routeRepo repository.RouteRepository,
	stopRepo repository.StopRepository,
) UserPersonalizedService {
	return &userPersonalizedService{
		userFavoriteRepo:   userFavoriteRepo,
		userPlaceRepo:      userPlaceRepo,
		userRecentViewRepo: userRecentViewRepo,
		userSavedNavRepo:   userSavedNavRepo,
		userRepo:           userRepo,
		routeRepo:          routeRepo,
		stopRepo:           stopRepo,
	}
}

// =====================
// Favorites
// =====================

func (s *userPersonalizedService) AddFavoriteRoute(ctx context.Context, userID, routeID uint) error {
	// Verify route exists
	_, err := s.routeRepo.FindByID(routeID)
	if err != nil {
		return errors.New("route not found")
	}

	// Check if already exists
	exists, err := s.userFavoriteRepo.Exists(userID, models.FavoriteTypeRoute, &routeID, nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrFavoriteAlreadyExists
	}

	favorite := &models.UserFavorite{
		UserID:       userID,
		FavoriteType: models.FavoriteTypeRoute,
		RouteID:      &routeID,
		CreatedAt:    time.Now(),
	}

	return s.userFavoriteRepo.Create(favorite)
}

func (s *userPersonalizedService) AddFavoriteStop(ctx context.Context, userID, stopID uint) error {
	// Verify stop exists
	_, err := s.stopRepo.FindByID(stopID)
	if err != nil {
		return errors.New("stop not found")
	}

	// Check if already exists
	exists, err := s.userFavoriteRepo.Exists(userID, models.FavoriteTypeStop, nil, &stopID)
	if err != nil {
		return err
	}
	if exists {
		return ErrFavoriteAlreadyExists
	}

	favorite := &models.UserFavorite{
		UserID:       userID,
		FavoriteType: models.FavoriteTypeStop,
		StopID:       &stopID,
		CreatedAt:    time.Now(),
	}

	return s.userFavoriteRepo.Create(favorite)
}

func (s *userPersonalizedService) RemoveFavoriteRoute(ctx context.Context, userID, routeID uint) error {
	return s.userFavoriteRepo.Delete(userID, models.FavoriteTypeRoute, &routeID, nil)
}

func (s *userPersonalizedService) RemoveFavoriteStop(ctx context.Context, userID, stopID uint) error {
	return s.userFavoriteRepo.Delete(userID, models.FavoriteTypeStop, nil, &stopID)
}

func (s *userPersonalizedService) GetFavoriteRoutes(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteRouteListResponse, error) {
	offset := (page - 1) * limit
	favorites, total, err := s.userFavoriteRepo.FindByUserID(userID, models.FavoriteTypeRoute, offset, limit)
	if err != nil {
		return nil, err
	}

	response := &dto.UserFavoriteRouteListResponse{
		Data:  make([]dto.UserFavoriteRouteResponse, 0, len(favorites)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, fav := range favorites {
		if fav.Route != nil {
			response.Data = append(response.Data, dto.UserFavoriteRouteResponse{
				ID:        fav.ID,
				Route:     *fav.Route,
				CreatedAt: fav.CreatedAt,
			})
		}
	}

	return response, nil
}

func (s *userPersonalizedService) GetFavoriteStops(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteStopListResponse, error) {
	offset := (page - 1) * limit
	favorites, total, err := s.userFavoriteRepo.FindByUserID(userID, models.FavoriteTypeStop, offset, limit)
	if err != nil {
		return nil, err
	}

	response := &dto.UserFavoriteStopListResponse{
		Data:  make([]dto.UserFavoriteStopResponse, 0, len(favorites)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, fav := range favorites {
		if fav.Stop != nil {
			response.Data = append(response.Data, dto.UserFavoriteStopResponse{
				ID:        fav.ID,
				Stop:      *fav.Stop,
				CreatedAt: fav.CreatedAt,
			})
		}
	}

	return response, nil
}

func (s *userPersonalizedService) IsFavoriteRoute(ctx context.Context, userID, routeID uint) (bool, error) {
	return s.userFavoriteRepo.Exists(userID, models.FavoriteTypeRoute, &routeID, nil)
}

func (s *userPersonalizedService) IsFavoriteStop(ctx context.Context, userID, stopID uint) (bool, error) {
	return s.userFavoriteRepo.Exists(userID, models.FavoriteTypeStop, nil, &stopID)
}

// =====================
// Places
// =====================

func (s *userPersonalizedService) CreatePlace(ctx context.Context, userID uint, req *dto.CreateUserPlaceRequest) (*dto.UserPlaceResponse, error) {
	place := &models.UserPlace{
		UserID:    userID,
		PlaceType: req.PlaceType,
		Name:      req.Name,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Address != nil {
		place.Address = *req.Address
	}
	if req.Notes != nil {
		place.Notes = *req.Notes
	}

	if req.PlaceType == models.PlaceTypeHome || req.PlaceType == models.PlaceTypeOffice {
		existingDefault, _ := s.userPlaceRepo.FindDefaultByUserID(userID, req.PlaceType)
		if existingDefault == nil {
			place.IsDefault = true
		}
	}

	err := s.userPlaceRepo.Create(place)
	if err != nil {
		return nil, err
	}

	return s.placeToResponse(place), nil
}

func (s *userPersonalizedService) UpdatePlace(ctx context.Context, userID, placeID uint, req *dto.UpdateUserPlaceRequest) (*dto.UserPlaceResponse, error) {
	place, err := s.userPlaceRepo.FindByID(placeID)
	if err != nil {
		return nil, ErrPlaceNotFound
	}
	if place.UserID != userID {
		return nil, ErrUnauthorizedPlaceAccess
	}

	// Update fields
	if req.PlaceType != nil {
		place.PlaceType = *req.PlaceType
	}
	if req.Name != nil {
		place.Name = *req.Name
	}
	if req.Latitude != nil {
		place.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		place.Longitude = *req.Longitude
	}
	if req.Address != nil {
		place.Address = *req.Address
	}
	if req.Notes != nil {
		place.Notes = *req.Notes
	}
	if req.IsDefault != nil && *req.IsDefault {
		if err := s.userPlaceRepo.SetDefault(userID, placeID, place.PlaceType); err != nil {
			return nil, err
		}
		place.IsDefault = true
	}

	place.UpdatedAt = time.Now()

	if err := s.userPlaceRepo.Update(place); err != nil {
		return nil, err
	}

	return s.placeToResponse(place), nil
}

func (s *userPersonalizedService) DeletePlace(ctx context.Context, userID, placeID uint) error {
	place, err := s.userPlaceRepo.FindByID(placeID)
	if err != nil {
		return ErrPlaceNotFound
	}
	if place.UserID != userID {
		return ErrUnauthorizedPlaceAccess
	}
	return s.userPlaceRepo.Delete(placeID)
}

func (s *userPersonalizedService) GetPlaces(ctx context.Context, userID uint) (*dto.UserPlaceListResponse, error) {
	places, err := s.userPlaceRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	response := &dto.UserPlaceListResponse{
		Data:  make([]dto.UserPlaceResponse, 0, len(places)),
		Total: int64(len(places)),
	}

	for _, place := range places {
		response.Data = append(response.Data, *s.placeToResponse(&place))
	}

	return response, nil
}

func (s *userPersonalizedService) GetPlaceByID(ctx context.Context, userID, placeID uint) (*dto.UserPlaceResponse, error) {
	place, err := s.userPlaceRepo.FindByID(placeID)
	if err != nil {
		return nil, ErrPlaceNotFound
	}
	if place.UserID != userID {
		return nil, ErrUnauthorizedPlaceAccess
	}
	return s.placeToResponse(place), nil
}

func (s *userPersonalizedService) placeToResponse(place *models.UserPlace) *dto.UserPlaceResponse {
	resp := &dto.UserPlaceResponse{
		ID:        place.ID,
		PlaceType: place.PlaceType,
		Name:      place.Name,
		Latitude:  place.Latitude,
		Longitude: place.Longitude,
		IsDefault: place.IsDefault,
		CreatedAt: place.CreatedAt,
		UpdatedAt: place.UpdatedAt,
	}
	if place.Address != "" {
		resp.Address = place.Address
	}
	if place.Notes != "" {
		resp.Notes = place.Notes
	}
	return resp
}

// =====================
// Recent Views
// =====================

func (s *userPersonalizedService) RecordRecentView(ctx context.Context, userID uint, req *dto.RecordRecentViewRequest) error {
	// Validate the request based on view type
	if req.ViewType == models.RecentViewTypeRoute && req.RouteID == nil {
		return errors.New("route_id is required for route view type")
	}
	if req.ViewType == models.RecentViewTypeStop && req.FromStopID == nil {
		return errors.New("from_stop_id is required for stop view type")
	}
	if req.ViewType == models.RecentViewTypeNavigation && (req.FromStopID == nil || req.ToStopID == nil) {
		return errors.New("from_stop_id and to_stop_id are required for navigation view type")
	}

	view := &models.UserRecentView{
		UserID:     userID,
		ViewType:   req.ViewType,
		RouteID:    req.RouteID,
		FromStopID: req.FromStopID,
		ToStopID:   req.ToStopID,
		ViewedAt:   time.Now(),
	}

	return s.userRecentViewRepo.Upsert(view)
}

func (s *userPersonalizedService) GetRecentRoutes(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteRouteListResponse, error) {
	offset := (page - 1) * limit
	views, total, err := s.userRecentViewRepo.FindByUserID(userID, models.RecentViewTypeRoute, offset, limit)
	if err != nil {
		return nil, err
	}

	response := &dto.UserFavoriteRouteListResponse{
		Data:  make([]dto.UserFavoriteRouteResponse, 0, len(views)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, view := range views {
		if view.Route != nil {
			response.Data = append(response.Data, dto.UserFavoriteRouteResponse{
				ID:        view.ID,
				Route:     *view.Route,
				CreatedAt: view.ViewedAt,
			})
		}
	}

	return response, nil
}

func (s *userPersonalizedService) GetRecentStops(ctx context.Context, userID uint, page, limit int) (*dto.UserFavoriteStopListResponse, error) {
	offset := (page - 1) * limit
	views, total, err := s.userRecentViewRepo.FindByUserID(userID, models.RecentViewTypeStop, offset, limit)
	if err != nil {
		return nil, err
	}

	response := &dto.UserFavoriteStopListResponse{
		Data:  make([]dto.UserFavoriteStopResponse, 0, len(views)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, view := range views {
		if view.FromStop != nil {
			response.Data = append(response.Data, dto.UserFavoriteStopResponse{
				ID:        view.ID,
				Stop:      *view.FromStop,
				CreatedAt: view.ViewedAt,
			})
		}
	}

	return response, nil
}

func (s *userPersonalizedService) GetRecentNavigations(ctx context.Context, userID uint, page, limit int) (*dto.UserSavedNavigationListResponse, error) {
	offset := (page - 1) * limit
	views, total, err := s.userRecentViewRepo.FindByUserID(userID, models.RecentViewTypeNavigation, offset, limit)
	if err != nil {
		return nil, err
	}

	response := &dto.UserSavedNavigationListResponse{
		Data:  make([]dto.UserSavedNavigationResponse, 0, len(views)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, view := range views {
		response.Data = append(response.Data, dto.UserSavedNavigationResponse{
			ID:   view.ID,
			From: s.stopToNavigationPoint(view.FromStop, nil),
			To:   s.stopToNavigationPoint(view.ToStop, nil),
		})
	}

	return response, nil
}

// =====================
// Saved Navigations
// =====================

func (s *userPersonalizedService) CreateSavedNavigation(ctx context.Context, userID uint, req *dto.CreateUserSavedNavigationRequest) (*dto.UserSavedNavigationResponse, error) {
	// Validate: must have either place_id or stop_id for both from and to
	if req.FromPlaceID == nil && req.FromStopID == nil {
		return nil, ErrInvalidNavigationPoint
	}
	if req.ToPlaceID == nil && req.ToStopID == nil {
		return nil, ErrInvalidNavigationPoint
	}

	nav := &models.UserSavedNavigation{
		UserID:        userID,
		Name:          s.getNavigationName(req),
		FromPlaceID:   req.FromPlaceID,
		FromStopID:    req.FromStopID,
		ToPlaceID:     req.ToPlaceID,
		ToStopID:      req.ToStopID,
		FromPlaceType: req.FromPlaceType,
		ToPlaceType:   req.ToPlaceType,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if req.FromPlaceType != nil {
		nav.FromPlaceType = req.FromPlaceType
	}
	if req.ToPlaceType != nil {
		nav.ToPlaceType = req.ToPlaceType
	}

	err := s.userSavedNavRepo.Create(nav)
	if err != nil {
		return nil, err
	}

	return s.savedNavToResponse(nav), nil
}

func (s *userPersonalizedService) UpdateSavedNavigation(ctx context.Context, userID, navID uint, req *dto.UpdateUserSavedNavigationRequest) (*dto.UserSavedNavigationResponse, error) {
	nav, err := s.userSavedNavRepo.FindByID(navID)
	if err != nil {
		return nil, ErrSavedNavigationNotFound
	}
	if nav.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Update fields
	if req.Name != nil {
		nav.Name = *req.Name
	}
	if req.FromPlaceType != nil {
		nav.FromPlaceType = req.FromPlaceType
	}
	if req.FromPlaceID != nil {
		nav.FromPlaceID = req.FromPlaceID
	}
	if req.FromStopID != nil {
		nav.FromStopID = req.FromStopID
	}
	if req.ToPlaceType != nil {
		nav.ToPlaceType = req.ToPlaceType
	}
	if req.ToPlaceID != nil {
		nav.ToPlaceID = req.ToPlaceID
	}
	if req.ToStopID != nil {
		nav.ToStopID = req.ToStopID
	}

	nav.UpdatedAt = time.Now()

	if err := s.userSavedNavRepo.Update(nav); err != nil {
		return nil, err
	}

	return s.savedNavToResponse(nav), nil
}

func (s *userPersonalizedService) DeleteSavedNavigation(ctx context.Context, userID, navID uint) error {
	nav, err := s.userSavedNavRepo.FindByID(navID)
	if err != nil {
		return ErrSavedNavigationNotFound
	}
	if nav.UserID != userID {
		return ErrUnauthorizedAccess
	}
	return s.userSavedNavRepo.Delete(navID)
}

func (s *userPersonalizedService) GetSavedNavigations(ctx context.Context, userID uint, page, limit int) (*dto.UserSavedNavigationListResponse, error) {
	offset := (page - 1) * limit
	navigations, total, err := s.userSavedNavRepo.FindByUserID(userID, offset, limit)
	if err != nil {
		return nil, err
	}

	response := &dto.UserSavedNavigationListResponse{
		Data:  make([]dto.UserSavedNavigationResponse, 0, len(navigations)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, nav := range navigations {
		response.Data = append(response.Data, *s.savedNavToResponse(&nav))
	}

	return response, nil
}

func (s *userPersonalizedService) savedNavToResponse(nav *models.UserSavedNavigation) *dto.UserSavedNavigationResponse {
	return &dto.UserSavedNavigationResponse{
		ID:        nav.ID,
		Name:      nav.Name,
		From:      s.placeOrStopToNavigationPoint(nav.FromPlace, nav.FromStop, nav.FromPlaceType),
		To:        s.placeOrStopToNavigationPoint(nav.ToPlace, nav.ToStop, nav.ToPlaceType),
		CreatedAt: nav.CreatedAt,
		UpdatedAt: nav.UpdatedAt,
	}
}

func (s *userPersonalizedService) placeOrStopToNavigationPoint(place *models.UserPlace, stop *models.Stop, placeType *string) dto.NavigationPoint {
	np := dto.NavigationPoint{}
	if place != nil {
		np.PlaceID = &place.ID
		np.PlaceName = place.Name
		np.PlaceType = string(place.PlaceType)
	}
	if stop != nil {
		np.StopID = &stop.ID
		np.StopName = stop.Name
	}
	if placeType != nil && *placeType != "" {
		np.PlaceType = *placeType
	}
	return np
}

func (s *userPersonalizedService) stopToNavigationPoint(stop *models.Stop, placeType *string) dto.NavigationPoint {
	if stop == nil {
		return dto.NavigationPoint{}
	}
	return dto.NavigationPoint{
		StopID:   &stop.ID,
		StopName: stop.Name,
	}
}

func (s *userPersonalizedService) getNavigationName(req *dto.CreateUserSavedNavigationRequest) string {
	if req.Name != nil && *req.Name != "" {
		return *req.Name
	}
	return ""
}

// =====================
// Analytics
// =====================

func (s *userPersonalizedService) GetUserAnalytics(ctx context.Context, userID uint) (*dto.UserAnalyticsResponse, error) {
	mostFrequentRoute, err := s.userRepo.GetMostFrequentRoute(userID)
	if err != nil {
		return nil, err
	}

	mostFrequentStop, err := s.userRepo.GetMostFrequentStop(userID)
	if err != nil {
		return nil, err
	}

	totalCheckIns, err := s.userRepo.GetTotalCheckInsByUserID(userID)
	if err != nil {
		return nil, err
	}

	totalRoutesTraveled, err := s.userRepo.GetTotalRoutesTraveledByUserID(userID)
	if err != nil {
		return nil, err
	}

	totalUniqueRoutes, err := s.userRepo.GetTotalUniqueRoutesByUserID(userID)
	if err != nil {
		return nil, err
	}

	totalDuration, err := s.userRepo.GetTotalCheckInDuration(userID)
	if err != nil {
		return nil, err
	}

	favoritePlacesCount, err := s.userPlaceRepo.CountByUserID(userID)
	if err != nil {
		return nil, err
	}

	favoriteRoutesCount, _ := s.userFavoriteRepo.CountByUserID(userID, ptr(models.FavoriteTypeRoute))
	favoriteStopsCount, _ := s.userFavoriteRepo.CountByUserID(userID, ptr(models.FavoriteTypeStop))

	savedNavigationsCount, err := s.userSavedNavRepo.CountByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserAnalyticsResponse{
		MostFrequentRoute:     mostFrequentRoute,
		MostFrequentStop:      mostFrequentStop,
		TotalCheckIns:         totalCheckIns,
		TotalRoutesTraveled:   totalRoutesTraveled,
		TotalUniqueRoutes:     totalUniqueRoutes,
		TotalDurationSeconds:  totalDuration,
		FavoritePlacesCount:   int(favoritePlacesCount),
		FavoriteRoutesCount:   int(favoriteRoutesCount),
		FavoriteStopsCount:    int(favoriteStopsCount),
		SavedNavigationsCount: int(savedNavigationsCount),
	}, nil
}

func ptr[T any](v T) *T {
	return &v
}
