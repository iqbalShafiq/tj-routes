package service

import (
	"tj-routes/internal/repository"
)

type ReputationService interface {
	AddPoints(userID uint, points int) error
	CalculateLevel(points int) string
	UpdateUserLevel(userID uint) error
	GetUserReputation(userID uint) (int, string, error)
}

type reputationService struct {
	userRepo repository.UserRepository
}

func NewReputationService(userRepo repository.UserRepository) ReputationService {
	return &reputationService{
		userRepo: userRepo,
	}
}

// CalculateLevel determines user level based on reputation points
func (s *reputationService) CalculateLevel(points int) string {
	switch {
	case points >= 1000:
		return "legend"
	case points >= 500:
		return "expert"
	case points >= 200:
		return "trusted"
	case points >= 50:
		return "contributor"
	default:
		return "newcomer"
	}
}

// AddPoints adds points to a user and updates their level
func (s *reputationService) AddPoints(userID uint, points int) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.ReputationPoints += points
	user.Level = s.CalculateLevel(user.ReputationPoints)

	return s.userRepo.Update(user)
}

// UpdateUserLevel recalculates and updates user level based on current points
func (s *reputationService) UpdateUserLevel(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.Level = s.CalculateLevel(user.ReputationPoints)
	return s.userRepo.Update(user)
}

// GetUserReputation returns user's current points and level
func (s *reputationService) GetUserReputation(userID uint) (int, string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, "", err
	}

	return user.ReputationPoints, user.Level, nil
}

