package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/utils"

	"gorm.io/gorm"
)

type UserService interface {
	Register(email, username, password string) (*models.User, error)
	Login(email, password string) (*models.User, string, string, error)
	AuthenticateOAuth(provider models.OAuthProvider, oauthID, email, name string) (*models.User, string, string, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	ListUsers(offset, limit int) ([]models.User, int64, error)
	UpdateUserRole(userID uint, role models.UserRole) error
	GetSystemUser() (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	config   *config.Config
	cache    cache.Cache
}

func NewUserService(userRepo repository.UserRepository, cfg *config.Config, cacheInstance cache.Cache) UserService {
	return &userService{
		userRepo: userRepo,
		config:   cfg,
		cache:    cacheInstance,
	}
}

func (s *userService) Register(email, username, password string) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:    email,
		Username: username,
		Password: &hashedPassword,
		Role:     models.RoleCommonUser,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) Login(email, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errors.New("invalid email or password")
		}
		return nil, "", "", fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user has password (not OAuth-only user)
	if user.Password == nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	// Verify password
	if !utils.CheckPasswordHash(password, *user.Password) {
		return nil, "", "", errors.New("invalid email or password")
	}

	// Generate tokens
	accessToken, err := utils.GenerateToken(user, &s.config.JWT)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user, &s.config.JWT)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *userService) AuthenticateOAuth(provider models.OAuthProvider, oauthID, email, name string) (*models.User, string, string, error) {
	// Try to find existing OAuth user
	user, err := s.userRepo.FindByOAuth(provider, oauthID)
	if err == nil {
		// User exists, generate tokens
		accessToken, err := utils.GenerateToken(user, &s.config.JWT)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
		}

		refreshToken, err := utils.GenerateRefreshToken(user, &s.config.JWT)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
		}

		return user, accessToken, refreshToken, nil
	}

	// Check if user exists with same email (link accounts)
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		// Link OAuth to existing account
		existingUser.OAuthProvider = &provider
		oauthIDStr := oauthID
		existingUser.OAuthID = &oauthIDStr
		if err := s.userRepo.Update(existingUser); err != nil {
			return nil, "", "", fmt.Errorf("failed to link OAuth account: %w", err)
		}

		accessToken, err := utils.GenerateToken(existingUser, &s.config.JWT)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
		}

		refreshToken, err := utils.GenerateRefreshToken(existingUser, &s.config.JWT)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
		}

		return existingUser, accessToken, refreshToken, nil
	}

	// Create new user
	user = &models.User{
		Email:         email,
		Username:      name,
		OAuthProvider: &provider,
		OAuthID:       &oauthID,
		Role:          models.RoleCommonUser,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := utils.GenerateToken(user, &s.config.JWT)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user, &s.config.JWT)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *userService) ListUsers(offset, limit int) ([]models.User, int64, error) {
	return s.userRepo.List(offset, limit)
}

func (s *userService) UpdateUserRole(userID uint, role models.UserRole) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.Role = role
	return s.userRepo.Update(user)
}

func (s *userService) GetSystemUser() (*models.User, error) {
	const systemUserEmail = "system@tj-routes.local"
	ctx := context.Background()
	key := cache.SystemUserKey()

	// Try to get from cache
	var user models.User
	if err := s.cache.Get(ctx, key, &user); err == nil {
		return &user, nil
	}

	// Cache miss, try to find existing system user
	userPtr, err := s.userRepo.FindByEmail(systemUserEmail)
	if err == nil {
		// Store in cache with long TTL
		ttl := time.Duration(s.config.Cache.SystemUserTTL) * time.Minute
		s.cache.Set(ctx, key, *userPtr, ttl)
		return userPtr, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find system user: %w", err)
	}

	// System user doesn't exist, create it
	userPtr = &models.User{
		Email:    systemUserEmail,
		Username: "system",
		Role:     models.RoleCommonUser,
		Password: nil, // No password for system user
	}

	if err := s.userRepo.Create(userPtr); err != nil {
		return nil, fmt.Errorf("failed to create system user: %w", err)
	}

	// Store in cache with long TTL
	ttl := time.Duration(s.config.Cache.SystemUserTTL) * time.Minute
	s.cache.Set(ctx, key, *userPtr, ttl)

	return userPtr, nil
}

