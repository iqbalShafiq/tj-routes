package service

import (
	"errors"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
)

type ChatRequestService interface {
	CreateChatRequest(senderID, receiverID uint, message string) (*models.ChatRequest, error)
	ListSentRequests(userID uint, offset, limit int) ([]models.ChatRequest, int64, error)
	ListReceivedRequests(userID uint, offset, limit int) ([]models.ChatRequest, int64, error)
	AcceptRequest(id, userID uint) error
	RejectRequest(id, userID uint) error
}

type chatRequestService struct {
	chatRequestRepo     repository.ChatRequestRepository
	conversationRepo    repository.ConversationRepository
	conversationService ConversationService
	userFollowRepo      repository.UserFollowRepository
	userRepo            repository.UserRepository
	cache               cache.Cache
	cfg                 *config.Config
}

func NewChatRequestService(
	chatRequestRepo repository.ChatRequestRepository,
	conversationRepo repository.ConversationRepository,
	conversationService ConversationService,
	userFollowRepo repository.UserFollowRepository,
	userRepo repository.UserRepository,
	cache cache.Cache,
	cfg *config.Config,
) ChatRequestService {
	return &chatRequestService{
		chatRequestRepo:     chatRequestRepo,
		conversationRepo:    conversationRepo,
		conversationService: conversationService,
		userFollowRepo:      userFollowRepo,
		userRepo:            userRepo,
		cache:               cache,
		cfg:                 cfg,
	}
}

func (s *chatRequestService) CreateChatRequest(senderID, receiverID uint, message string) (*models.ChatRequest, error) {
	if senderID == receiverID {
		return nil, errors.New("cannot send chat request to yourself")
	}

	_, err := s.userRepo.FindByID(receiverID)
	if err != nil {
		return nil, errors.New("receiver not found")
	}

	isFollowing, err := s.userFollowRepo.IsFollowing(senderID, receiverID)
	if err != nil {
		return nil, err
	}
	if isFollowing {
		return nil, errors.New("users are already following each other")
	}

	existingRequest, _ := s.chatRequestRepo.FindBySenderAndReceiver(senderID, receiverID)
	if existingRequest != nil && existingRequest.Status == "pending" {
		return nil, errors.New("pending chat request already exists")
	}

	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	if len(message) > 500 {
		return nil, errors.New("message too long")
	}

	chatRequest := &models.ChatRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Message:    message,
		Status:     "pending",
	}

	if err := s.chatRequestRepo.Create(chatRequest); err != nil {
		return nil, err
	}

	return s.chatRequestRepo.FindByID(chatRequest.ID)
}

func (s *chatRequestService) ListSentRequests(userID uint, offset, limit int) ([]models.ChatRequest, int64, error) {
	requests, total, err := s.chatRequestRepo.ListSent(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (s *chatRequestService) ListReceivedRequests(userID uint, offset, limit int) ([]models.ChatRequest, int64, error) {
	requests, total, err := s.chatRequestRepo.ListReceived(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (s *chatRequestService) AcceptRequest(id, userID uint) error {
	request, err := s.chatRequestRepo.FindByID(id)
	if err != nil {
		return errors.New("chat request not found")
	}

	if request.ReceiverID != userID {
		return errors.New("not authorized to accept this request")
	}

	if request.Status != "pending" {
		return errors.New("request has already been processed")
	}

	_, err = s.conversationService.GetOrCreateConversation(request.SenderID, request.ReceiverID)
	if err != nil {
		return err
	}

	if err := s.chatRequestRepo.UpdateStatus(id, "accepted"); err != nil {
		return err
	}

	return nil
}

func (s *chatRequestService) RejectRequest(id, userID uint) error {
	request, err := s.chatRequestRepo.FindByID(id)
	if err != nil {
		return errors.New("chat request not found")
	}

	if request.ReceiverID != userID {
		return errors.New("not authorized to reject this request")
	}

	if request.Status != "pending" {
		return errors.New("request has already been processed")
	}

	return s.chatRequestRepo.UpdateStatus(id, "rejected")
}
