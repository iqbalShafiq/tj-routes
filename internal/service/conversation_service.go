package service

import (
	"context"
	"errors"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ConversationService interface {
	GetOrCreateConversation(userID1, userID2 uint) (*models.Conversation, error)
	GetConversationByID(id uint) (*models.Conversation, error)
	ListUserConversations(userID uint, offset, limit int) ([]models.Conversation, int64, error)
	DeleteConversation(id, userID uint) error
	GetMessages(conversationID, userID uint, offset, limit int) ([]models.Message, int64, error)
	MarkAsRead(conversationID, userID uint) error
	AddParticipant(conversationID, userID, requestingUserID uint) error
	RemoveParticipant(conversationID, userID uint) error
}

type conversationService struct {
	conversationRepo repository.ConversationRepository
	messageRepo      repository.MessageRepository
	userFollowRepo   repository.UserFollowRepository
	userRepo         repository.UserRepository
	cache            cache.Cache
	cfg              *config.Config
}

func NewConversationService(
	conversationRepo repository.ConversationRepository,
	messageRepo repository.MessageRepository,
	userFollowRepo repository.UserFollowRepository,
	userRepo repository.UserRepository,
	cache cache.Cache,
	cfg *config.Config,
) ConversationService {
	return &conversationService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		userFollowRepo:   userFollowRepo,
		userRepo:         userRepo,
		cache:            cache,
		cfg:              cfg,
	}
}

func (s *conversationService) GetOrCreateConversation(userID1, userID2 uint) (*models.Conversation, error) {
	if userID1 == userID2 {
		return nil, errors.New("cannot create conversation with yourself")
	}

	conversation, err := s.conversationRepo.FindByUsers(userID1, userID2)
	if err == nil {
		return conversation, nil
	}

	conversation = &models.Conversation{
		Type: "direct",
	}

	if err := s.conversationRepo.Create(conversation); err != nil {
		return nil, err
	}

	now := time.Now()

	participant1 := &models.ConversationParticipant{
		ConversationID: conversation.ID,
		UserID:         userID1,
		LastReadAt:     &now,
		JoinedAt:       now,
	}

	participant2 := &models.ConversationParticipant{
		ConversationID: conversation.ID,
		UserID:         userID2,
		LastReadAt:     &now,
		JoinedAt:       now,
	}

	if err := s.conversationRepo.AddParticipant(participant1); err != nil {
		return nil, err
	}
	if err := s.conversationRepo.AddParticipant(participant2); err != nil {
		return nil, err
	}

	return s.GetConversationByID(conversation.ID)
}

func (s *conversationService) GetConversationByID(id uint) (*models.Conversation, error) {
	if s.cache != nil {
		cacheKey := cache.ConversationKey(id)
		var conversation models.Conversation
		if err := s.cache.Get(context.Background(), cacheKey, &conversation); err == nil {
			return &conversation, nil
		}
	}

	conversation, err := s.conversationRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		cacheKey := cache.ConversationKey(id)
		s.cache.Set(context.Background(), cacheKey, conversation, 30*time.Minute)
	}

	return conversation, nil
}

func (s *conversationService) ListUserConversations(userID uint, offset, limit int) ([]models.Conversation, int64, error) {
	conversations, total, err := s.conversationRepo.ListByUserID(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

func (s *conversationService) DeleteConversation(id, userID uint) error {
	conversation, err := s.conversationRepo.FindByID(id)
	if err != nil {
		return errors.New("conversation not found")
	}

	// Verify user is a participant
	participants, err := s.conversationRepo.FindParticipants(id)
	if err != nil {
		return err
	}

	isParticipant := false
	for _, p := range participants {
		if p.UserID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return errors.New("not authorized to delete this conversation")
	}

	// Only allow deletion of direct conversations or if user is the creator
	if conversation.Type != "direct" {
		return errors.New("only direct conversations can be deleted by participants")
	}

	if err := s.conversationRepo.Delete(id); err != nil {
		return err
	}

	if s.cache != nil {
		s.cache.Delete(context.Background(), cache.ConversationKey(id))
	}

	return nil
}

func (s *conversationService) GetMessages(conversationID, userID uint, offset, limit int) ([]models.Message, int64, error) {
	_, err := s.conversationRepo.FindByID(conversationID)
	if err != nil {
		return nil, 0, errors.New("conversation not found")
	}

	messages, total, err := s.messageRepo.ListByConversationID(conversationID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (s *conversationService) MarkAsRead(conversationID, userID uint) error {
	_, err := s.conversationRepo.FindByID(conversationID)
	if err != nil {
		return errors.New("conversation not found")
	}

	participants, err := s.conversationRepo.FindParticipants(conversationID)
	if err != nil {
		return err
	}

	isParticipant := false
	for _, p := range participants {
		if p.UserID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return errors.New("not a participant in this conversation")
	}

	now := time.Now()
	if err := s.conversationRepo.UpdateLastRead(conversationID, userID, now); err != nil {
		return err
	}

	if s.cache != nil {
		s.cache.Delete(context.Background(), cache.UnreadCountKey(userID))
	}

	return nil
}

func (s *conversationService) AddParticipant(conversationID, userID, requestingUserID uint) error {
	conversation, err := s.conversationRepo.FindByID(conversationID)
	if err != nil {
		return errors.New("conversation not found")
	}

	if conversation.Type == "direct" {
		return errors.New("cannot add participants to direct conversations")
	}

	// Verify requesting user is a participant
	participants, err := s.conversationRepo.FindParticipants(conversationID)
	if err != nil {
		return err
	}

	isRequestingUserParticipant := false
	for _, p := range participants {
		if p.UserID == requestingUserID {
			isRequestingUserParticipant = true
		}
		if p.UserID == userID {
			return errors.New("user is already a participant")
		}
	}

	if !isRequestingUserParticipant {
		return errors.New("not authorized to add participants")
	}

	_, err = s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	now := time.Now()
	participant := &models.ConversationParticipant{
		ConversationID: conversationID,
		UserID:         userID,
		LastReadAt:     &now,
		JoinedAt:       now,
	}

	return s.conversationRepo.AddParticipant(participant)
}

func (s *conversationService) RemoveParticipant(conversationID, userID uint) error {
	conversation, err := s.conversationRepo.FindByID(conversationID)
	if err != nil {
		return errors.New("conversation not found")
	}

	if conversation.Type == "direct" {
		return errors.New("cannot remove participants from direct conversations")
	}

	return s.conversationRepo.RemoveParticipant(conversationID, userID)
}
