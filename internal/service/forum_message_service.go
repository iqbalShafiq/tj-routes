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

type ForumMessageService interface {
	CreateMessage(forumID uint, userID uint, content string) (*models.ForumMessage, error)
	GetMessageByID(id uint) (*models.ForumMessage, error)
	ListForumMessages(forumID uint, offset, limit int) ([]models.ForumMessage, int64, error)
	DeleteMessage(id uint, userID uint) error
	DeleteOldMessages(cutoff time.Time) error
}

type forumMessageService struct {
	forumMessageRepo repository.ForumMessageRepository
	forumMemberRepo  repository.ForumMemberRepository
	cache            cache.Cache
	config           *config.Config
}

func NewForumMessageService(
	forumMessageRepo repository.ForumMessageRepository,
	forumMemberRepo repository.ForumMemberRepository,
	cacheInstance cache.Cache,
	cfg *config.Config,
) ForumMessageService {
	return &forumMessageService{
		forumMessageRepo: forumMessageRepo,
		forumMemberRepo:  forumMemberRepo,
		cache:            cacheInstance,
		config:           cfg,
	}
}

const (
	maxForumMessageLength = 5000
	forumMessageCacheTTL  = 24 * time.Hour
)

func (s *forumMessageService) CreateMessage(forumID uint, userID uint, content string) (*models.ForumMessage, error) {
	if !s.config.Chat.ForumChatEnabled {
		return nil, errors.New("forum chat is disabled")
	}

	if len(content) == 0 {
		return nil, errors.New("message content cannot be empty")
	}

	if len(content) > maxForumMessageLength {
		return nil, errors.New("message content is too long")
	}

	isMember, err := s.forumMemberRepo.IsMember(forumID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not a member of this forum")
	}

	message := &models.ForumMessage{
		ForumID: forumID,
		UserID:  userID,
		Content: content,
	}

	if err := s.forumMessageRepo.Create(message); err != nil {
		return nil, err
	}

	if s.cache != nil {
		ctx := context.Background()
		cacheKey := cache.ForumChatMessagesKey(forumID)
		s.cache.Delete(ctx, cacheKey)
	}

	return message, nil
}

func (s *forumMessageService) GetMessageByID(id uint) (*models.ForumMessage, error) {
	return s.forumMessageRepo.FindByID(id)
}

func (s *forumMessageService) ListForumMessages(forumID uint, offset, limit int) ([]models.ForumMessage, int64, error) {
	if limit > s.config.Chat.ForumChatMaxMessages {
		limit = s.config.Chat.ForumChatMaxMessages
	}

	if s.cache != nil {
		cacheKey := cache.ForumChatMessagesKey(forumID)
		ctx := context.Background()

		var cachedMessages []models.ForumMessage
		if err := s.cache.Get(ctx, cacheKey, &cachedMessages); err == nil {
			if len(cachedMessages) > offset {
				end := offset + limit
				if end > len(cachedMessages) {
					end = len(cachedMessages)
				}
				return cachedMessages[offset:end], int64(len(cachedMessages)), nil
			}
		}
	}

	messages, total, err := s.forumMessageRepo.ListByForumID(forumID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	if s.cache != nil && offset == 0 && len(messages) > 0 {
		ctx := context.Background()
		cacheKey := cache.ForumChatMessagesKey(forumID)
		s.cache.Set(ctx, cacheKey, messages, forumMessageCacheTTL)
	}

	return messages, total, nil
}

func (s *forumMessageService) DeleteMessage(id uint, userID uint) error {
	message, err := s.forumMessageRepo.FindByID(id)
	if err != nil {
		return err
	}

	if message.UserID != userID {
		return errors.New("user is not the owner of this message")
	}

	if err := s.forumMessageRepo.Delete(id); err != nil {
		return err
	}

	if s.cache != nil {
		ctx := context.Background()
		cacheKey := cache.ForumChatMessagesKey(message.ForumID)
		s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

func (s *forumMessageService) DeleteOldMessages(cutoff time.Time) error {
	return s.forumMessageRepo.DeleteOlderThan(cutoff)
}
