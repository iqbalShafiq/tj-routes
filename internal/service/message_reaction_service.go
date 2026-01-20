package service

import (
	"errors"

	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type MessageReactionService interface {
	CreateReaction(messageID, userID uint, emoji string) (*models.MessageReaction, error)
	RemoveReaction(messageID, userID uint) error
	GetMessageReactions(messageID uint) ([]models.MessageReaction, error)
}

type messageReactionService struct {
	messageReactionRepo repository.MessageReactionRepository
	messageRepo         repository.MessageRepository
}

func NewMessageReactionService(
	messageReactionRepo repository.MessageReactionRepository,
	messageRepo repository.MessageRepository,
) MessageReactionService {
	return &messageReactionService{
		messageReactionRepo: messageReactionRepo,
		messageRepo:         messageRepo,
	}
}

func (s *messageReactionService) CreateReaction(messageID, userID uint, emoji string) (*models.MessageReaction, error) {
	if emoji == "" {
		return nil, errors.New("emoji cannot be empty")
	}

	_, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, errors.New("message not found")
	}

	existingReaction, err := s.messageReactionRepo.FindByUserAndMessage(userID, messageID)
	if err == nil {
		// If same emoji, toggle off (delete)
		if existingReaction.Emoji == emoji {
			if err := s.messageReactionRepo.Delete(existingReaction.ID); err != nil {
				return nil, err
			}
			return nil, nil
		}

		// Different emoji, delete old and create new reaction
		if err := s.messageReactionRepo.Delete(existingReaction.ID); err != nil {
			return nil, err
		}
		// Continue to create new reaction with different emoji below
	}

	reaction := &models.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	if err := s.messageReactionRepo.Create(reaction); err != nil {
		return nil, err
	}

	return s.messageReactionRepo.FindByID(reaction.ID)
}

func (s *messageReactionService) RemoveReaction(messageID, userID uint) error {
	_, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return errors.New("message not found")
	}

	return s.messageReactionRepo.DeleteByUserAndMessage(userID, messageID)
}

func (s *messageReactionService) GetMessageReactions(messageID uint) ([]models.MessageReaction, error) {
	_, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, errors.New("message not found")
	}

	return s.messageReactionRepo.ListByMessageID(messageID)
}
