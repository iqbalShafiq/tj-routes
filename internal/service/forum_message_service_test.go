package service

import (
	"errors"
	"testing"
	"time"

	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestForumMessageService_CreateMessage(t *testing.T) {
	mockForumMessageRepo := new(mocks.MockForumMessageRepository)
	mockForumMemberRepo := new(mocks.MockForumMemberRepository)
	cfg := &config.Config{
		Chat: config.ChatConfig{
			ForumChatEnabled:     true,
			ForumChatMaxMessages: 100,
		},
	}

	service := NewForumMessageService(mockForumMessageRepo, mockForumMemberRepo, nil, cfg)

	t.Run("success", func(t *testing.T) {
		forumID := uint(1)
		userID := uint(1)
		content := "Hello, world!"

		mockForumMemberRepo.On("IsMember", forumID, userID).Return(true, nil)
		mockForumMessageRepo.On("Create", mock.AnythingOfType("*models.ForumMessage")).Return(nil)

		message, err := service.CreateMessage(forumID, userID, content)

		assert.NoError(t, err)
		assert.NotNil(t, message)
		assert.Equal(t, forumID, message.ForumID)
		assert.Equal(t, userID, message.UserID)
		assert.Equal(t, content, message.Content)

		mockForumMemberRepo.AssertExpectations(t)
		mockForumMessageRepo.AssertExpectations(t)
	})

	t.Run("forum chat disabled", func(t *testing.T) {
		disabledCfg := &config.Config{Chat: config.ChatConfig{ForumChatEnabled: false}}
		serviceDisabled := NewForumMessageService(mockForumMessageRepo, mockForumMemberRepo, nil, disabledCfg)

		_, err := serviceDisabled.CreateMessage(1, 1, "test")

		assert.Error(t, err)
		assert.Equal(t, "forum chat is disabled", err.Error())
	})

	t.Run("empty content", func(t *testing.T) {
		_, err := service.CreateMessage(1, 1, "")

		assert.Error(t, err)
		assert.Equal(t, "message content cannot be empty", err.Error())
	})

	t.Run("content too long", func(t *testing.T) {
		longContent := string(make([]byte, 5001))

		_, err := service.CreateMessage(1, 1, longContent)

		assert.Error(t, err)
		assert.Equal(t, "message content is too long", err.Error())
	})

	t.Run("user not member", func(t *testing.T) {
		mockForumMemberRepo.On("IsMember", uint(1), uint(1)).Return(false, nil)

		_, err := service.CreateMessage(1, 1, "test")

		assert.Error(t, err)
		assert.Equal(t, "user is not a member of this forum", err.Error())
		mockForumMemberRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockForumMemberRepo.On("IsMember", uint(1), uint(1)).Return(true, nil)
		mockForumMessageRepo.On("Create", mock.AnythingOfType("*models.ForumMessage")).Return(errors.New("db error"))

		_, err := service.CreateMessage(1, 1, "test")

		assert.Error(t, err)
		mockForumMemberRepo.AssertExpectations(t)
		mockForumMessageRepo.AssertExpectations(t)
	})
}

func TestForumMessageService_GetMessageByID(t *testing.T) {
	mockForumMessageRepo := new(mocks.MockForumMessageRepository)
	mockForumMemberRepo := new(mocks.MockForumMemberRepository)
	cfg := &config.Config{
		Chat: config.ChatConfig{
			ForumChatEnabled:     true,
			ForumChatMaxMessages: 100,
		},
	}

	service := NewForumMessageService(mockForumMessageRepo, mockForumMemberRepo, nil, cfg)

	t.Run("success", func(t *testing.T) {
		expectedMessage := &models.ForumMessage{
			ID:      1,
			ForumID: 1,
			UserID:  1,
			Content: "test",
		}

		mockForumMessageRepo.On("FindByID", uint(1)).Return(expectedMessage, nil)

		message, err := service.GetMessageByID(1)

		assert.NoError(t, err)
		assert.Equal(t, expectedMessage, message)
		mockForumMessageRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockForumMessageRepo.On("FindByID", uint(999)).Return(nil, errors.New("not found"))

		_, err := service.GetMessageByID(999)

		assert.Error(t, err)
		mockForumMessageRepo.AssertExpectations(t)
	})
}

func TestForumMessageService_ListForumMessages(t *testing.T) {
	mockForumMessageRepo := new(mocks.MockForumMessageRepository)
	mockForumMemberRepo := new(mocks.MockForumMemberRepository)
	cfg := &config.Config{
		Chat: config.ChatConfig{
			ForumChatEnabled:     true,
			ForumChatMaxMessages: 100,
		},
	}

	service := NewForumMessageService(mockForumMessageRepo, mockForumMemberRepo, nil, cfg)

	t.Run("success", func(t *testing.T) {
		expectedMessages := []models.ForumMessage{
			{ID: 1, ForumID: 1, UserID: 1, Content: "message 1"},
			{ID: 2, ForumID: 1, UserID: 2, Content: "message 2"},
		}

		mockForumMessageRepo.On("ListByForumID", uint(1), 0, 50).Return(expectedMessages, int64(2), nil)

		messages, total, err := service.ListForumMessages(1, 0, 50)

		assert.NoError(t, err)
		assert.Equal(t, expectedMessages, messages)
		assert.Equal(t, int64(2), total)
		mockForumMessageRepo.AssertExpectations(t)
	})

	t.Run("limit capped at max", func(t *testing.T) {
		expectedMessages := []models.ForumMessage{}

		mockForumMessageRepo.On("ListByForumID", uint(1), 0, 100).Return(expectedMessages, int64(0), nil)

		messages, _, err := service.ListForumMessages(1, 0, 200)

		assert.NoError(t, err)
		assert.Equal(t, expectedMessages, messages)
		mockForumMessageRepo.AssertExpectations(t)
	})
}

func TestForumMessageService_DeleteMessage(t *testing.T) {
	mockForumMessageRepo := new(mocks.MockForumMessageRepository)
	mockForumMemberRepo := new(mocks.MockForumMemberRepository)
	cfg := &config.Config{
		Chat: config.ChatConfig{
			ForumChatEnabled:     true,
			ForumChatMaxMessages: 100,
		},
	}

	service := NewForumMessageService(mockForumMessageRepo, mockForumMemberRepo, nil, cfg)

	t.Run("success", func(t *testing.T) {
		existingMessage := &models.ForumMessage{
			ID:      1,
			ForumID: 1,
			UserID:  1,
			Content: "test",
		}

		mockForumMessageRepo.On("FindByID", uint(1)).Return(existingMessage, nil)
		mockForumMessageRepo.On("Delete", uint(1)).Return(nil)

		err := service.DeleteMessage(1, 1)

		assert.NoError(t, err)
		mockForumMessageRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockForumMessageRepo.On("FindByID", uint(999)).Return(nil, errors.New("not found"))

		err := service.DeleteMessage(999, 1)

		assert.Error(t, err)
		mockForumMessageRepo.AssertExpectations(t)
	})

	t.Run("not owner", func(t *testing.T) {
		existingMessage := &models.ForumMessage{
			ID:      1,
			ForumID: 1,
			UserID:  2,
			Content: "test",
		}

		mockForumMessageRepo.On("FindByID", uint(1)).Return(existingMessage, nil)

		err := service.DeleteMessage(1, 1)

		assert.Error(t, err)
		assert.Equal(t, "user is not the owner of this message", err.Error())
		mockForumMessageRepo.AssertExpectations(t)
	})
}

func TestForumMessageService_DeleteOldMessages(t *testing.T) {
	mockForumMessageRepo := new(mocks.MockForumMessageRepository)
	mockForumMemberRepo := new(mocks.MockForumMemberRepository)
	cfg := &config.Config{
		Chat: config.ChatConfig{
			ForumChatEnabled:     true,
			ForumChatMaxMessages: 100,
		},
	}

	service := NewForumMessageService(mockForumMessageRepo, mockForumMemberRepo, nil, cfg)

	t.Run("success", func(t *testing.T) {
		cutoff := time.Now().Add(-24 * time.Hour)
		mockForumMessageRepo.On("DeleteOlderThan", cutoff).Return(nil)

		err := service.DeleteOldMessages(cutoff)

		assert.NoError(t, err)
		mockForumMessageRepo.AssertExpectations(t)
	})
}
