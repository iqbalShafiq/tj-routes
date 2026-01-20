package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tj-routes/internal/cache"
)

const (
	typingTimeout  = 5 * time.Second
	typingDebounce = 500 * time.Millisecond
)

type TypingManager struct {
	cache      cache.Cache
	mu         sync.RWMutex
	typingData map[uint]map[uint]time.Time
}

type TypingEvent struct {
	UserID         uint  `json:"user_id"`
	ConversationID *uint `json:"conversation_id,omitempty"`
	GroupID        *uint `json:"group_id,omitempty"`
	IsTyping       bool  `json:"is_typing"`
}

func NewTypingManager(cache cache.Cache) *TypingManager {
	tm := &TypingManager{
		cache:      cache,
		typingData: make(map[uint]map[uint]time.Time),
	}

	go tm.cleanupExpiredTyping()

	return tm
}

func (tm *TypingManager) UserStartedTyping(ctx context.Context, userID uint, conversationID *uint, groupID *uint) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var roomID uint
	if conversationID != nil {
		roomID = *conversationID
	} else if groupID != nil {
		roomID = *groupID
	} else {
		return fmt.Errorf("either conversation_id or group_id must be provided")
	}

	if tm.typingData[roomID] == nil {
		tm.typingData[roomID] = make(map[uint]time.Time)
	}

	typingKey := cache.TypingIndicatorKey(roomID)
	if tm.cache != nil && tm.cache.IsAvailable(ctx) {
		if err := tm.cache.Set(ctx, typingKey, userID, typingTimeout); err != nil {
			return fmt.Errorf("failed to set typing indicator: %w", err)
		}
	}

	tm.typingData[roomID][userID] = time.Now().Add(typingTimeout)

	return nil
}

func (tm *TypingManager) UserStoppedTyping(ctx context.Context, userID uint, conversationID *uint, groupID *uint) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var roomID uint
	if conversationID != nil {
		roomID = *conversationID
	} else if groupID != nil {
		roomID = *groupID
	} else {
		return fmt.Errorf("either conversation_id or group_id must be provided")
	}

	if tm.typingData[roomID] != nil {
		delete(tm.typingData[roomID], userID)
		if len(tm.typingData[roomID]) == 0 {
			delete(tm.typingData, roomID)
		}
	}

	if tm.cache != nil && tm.cache.IsAvailable(ctx) {
		typingKey := cache.TypingIndicatorKey(roomID)
		tm.cache.Delete(ctx, typingKey)
	}

	return nil
}

func (tm *TypingManager) GetTypingUsers(ctx context.Context, conversationID *uint, groupID *uint) ([]uint, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var roomID uint
	if conversationID != nil {
		roomID = *conversationID
	} else if groupID != nil {
		roomID = *groupID
	} else {
		return nil, fmt.Errorf("either conversation_id or group_id must be provided")
	}

	typingUsers := make([]uint, 0)

	if tm.typingData[roomID] != nil {
		now := time.Now()
		for userID, expiry := range tm.typingData[roomID] {
			if now.Before(expiry) {
				typingUsers = append(typingUsers, userID)
			}
		}
	}

	return typingUsers, nil
}

func (tm *TypingManager) IsUserTyping(ctx context.Context, userID uint, conversationID *uint, groupID *uint) (bool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var roomID uint
	if conversationID != nil {
		roomID = *conversationID
	} else if groupID != nil {
		roomID = *groupID
	} else {
		return false, fmt.Errorf("either conversation_id or group_id must be provided")
	}

	if tm.typingData[roomID] == nil {
		return false, nil
	}

	expiry, exists := tm.typingData[roomID][userID]
	if !exists {
		return false, nil
	}

	return time.Now().Before(expiry), nil
}

func (tm *TypingManager) CreateTypingEvent(userID uint, conversationID *uint, groupID *uint, isTyping bool) WSMessage {
	return WSMessage{
		Type: "typing",
		Payload: TypingEvent{
			UserID:         userID,
			ConversationID: conversationID,
			GroupID:        groupID,
			IsTyping:       isTyping,
		},
	}
}

func (tm *TypingManager) BroadcastTypingEvent(hub *Hub, userID uint, conversationID *uint, groupID *uint, isTyping bool) {
	roomID := BuildRoomID(conversationID, groupID, nil)
	event := tm.CreateTypingEvent(userID, conversationID, groupID, isTyping)

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	hub.BroadcastToRoom(roomID, data)
}

func (tm *TypingManager) cleanupExpiredTyping() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tm.mu.Lock()
		now := time.Now()
		for roomID := range tm.typingData {
			for userID, expiry := range tm.typingData[roomID] {
				if now.After(expiry) {
					delete(tm.typingData[roomID], userID)
				}
			}
			if len(tm.typingData[roomID]) == 0 {
				delete(tm.typingData, roomID)
			}
		}
		tm.mu.Unlock()
	}
}

type TypingDebouncer struct {
	lastUpdate time.Time
	mu         sync.Mutex
}

func (td *TypingDebouncer) ShouldBroadcast() bool {
	td.mu.Lock()
	defer td.mu.Unlock()

	now := time.Now()
	if now.Sub(td.lastUpdate) < typingDebounce {
		return false
	}

	td.lastUpdate = now
	return true
}
