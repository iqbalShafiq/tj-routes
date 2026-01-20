package service

import (
	"errors"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type MessageService interface {
	CreateMessage(content string, senderID uint, conversationID *uint, groupID *uint) (*models.Message, error)
	GetMessageByID(id uint) (*models.Message, error)
	ListConversationMessages(conversationID uint, offset, limit int) ([]models.Message, int64, error)
	ListGroupMessages(groupID uint, offset, limit int) ([]models.Message, int64, error)
	DeleteMessage(id, userID uint) error
	UpdateMessageStatus(id, userID uint, status string) error
}

type messageService struct {
	messageRepo       repository.MessageRepository
	conversationRepo  repository.ConversationRepository
	groupRepo         repository.GroupChatRepository
	groupMemberRepo   repository.GroupMemberRepository
	cache             cache.Cache
	cfg               *config.Config
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	conversationRepo repository.ConversationRepository,
	groupRepo repository.GroupChatRepository,
	groupMemberRepo repository.GroupMemberRepository,
	cache cache.Cache,
	cfg *config.Config,
) MessageService {
	return &messageService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		groupRepo:        groupRepo,
		groupMemberRepo:  groupMemberRepo,
		cache:            cache,
		cfg:              cfg,
	}
}

func (s *messageService) CreateMessage(content string, senderID uint, conversationID *uint, groupID *uint) (*models.Message, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	if len(content) > s.cfg.Chat.MessageMaxLength {
		return nil, errors.New("message too long")
	}

	if (conversationID == nil && groupID == nil) || (conversationID != nil && groupID != nil) {
		return nil, errors.New("message must belong to either a conversation or a group")
	}

	if conversationID != nil {
		_, err := s.conversationRepo.FindByID(*conversationID)
		if err != nil {
			return nil, errors.New("conversation not found")
		}
	}

	if groupID != nil {
		_, err := s.groupRepo.FindByID(*groupID)
		if err != nil {
			return nil, errors.New("group not found")
		}
	}

	message := &models.Message{
		Content:        content,
		SenderID:       senderID,
		ConversationID: conversationID,
		GroupID:        groupID,
		MessageType:    "text",
		Status:         "sent",
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return s.messageRepo.FindByID(message.ID)
}

func (s *messageService) GetMessageByID(id uint) (*models.Message, error) {
	return s.messageRepo.FindByID(id)
}

func (s *messageService) ListConversationMessages(conversationID uint, offset, limit int) ([]models.Message, int64, error) {
	_, err := s.conversationRepo.FindByID(conversationID)
	if err != nil {
		return nil, 0, errors.New("conversation not found")
	}

	return s.messageRepo.ListByConversationID(conversationID, offset, limit)
}

func (s *messageService) ListGroupMessages(groupID uint, offset, limit int) ([]models.Message, int64, error) {
	_, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, 0, errors.New("group not found")
	}

	return s.messageRepo.ListByGroupID(groupID, offset, limit)
}

func (s *messageService) DeleteMessage(id, userID uint) error {
	message, err := s.messageRepo.FindByID(id)
	if err != nil {
		return errors.New("message not found")
	}

	if message.SenderID != userID {
		return errors.New("not authorized to delete this message")
	}

	return s.messageRepo.Delete(id)
}

func (s *messageService) UpdateMessageStatus(id, userID uint, status string) error {
	validStatuses := map[string]bool{
		"sent":      true,
		"delivered": true,
		"read":      true,
	}

	if !validStatuses[status] {
		return errors.New("invalid message status")
	}

	// Get the message to check authorization
	message, err := s.messageRepo.FindByID(id)
	if err != nil {
		return errors.New("message not found")
	}

	// User can update status if they're the recipient (not the sender)
	// For conversation messages, check if user is a participant
	if message.ConversationID != nil {
		participants, err := s.conversationRepo.FindParticipants(*message.ConversationID)
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
			return errors.New("not authorized to update message status")
		}
	}

	// For group messages, check if user is a member
	if message.GroupID != nil {
		_, err := s.groupMemberRepo.FindByGroupAndUser(*message.GroupID, userID)
		if err != nil {
			return errors.New("not authorized to update message status")
		}
	}

	return s.messageRepo.UpdateStatus(id, status)
}
