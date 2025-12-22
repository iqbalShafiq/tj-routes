package service

import (
	"errors"
	"fmt"

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
}

type userService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

func NewUserService(userRepo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		config:   cfg,
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

