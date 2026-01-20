package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"tj-routes/internal/cache"
)

const (
	presenceTTL = 5 * time.Minute
)

type PresenceStatus string

const (
	PresenceStatusOnline  PresenceStatus = "online"
	PresenceStatusOffline PresenceStatus = "offline"
)

type PresenceManager struct {
	cache cache.Cache
}

type PresenceEvent struct {
	UserID uint           `json:"user_id"`
	Status PresenceStatus `json:"status"`
	RoomID string         `json:"room_id"`
}

func NewPresenceManager(cache cache.Cache) *PresenceManager {
	return &PresenceManager{
		cache: cache,
	}
}

func (pm *PresenceManager) SetUserOnline(ctx context.Context, userID uint, roomID string) error {
	userKey := cache.UserPresenceKey(userID)

	if err := pm.cache.Set(ctx, userKey, PresenceStatusOnline, presenceTTL); err != nil {
		return fmt.Errorf("failed to set user presence: %w", err)
	}

	roomOnlineKey := fmt.Sprintf("chat:presence:room:%s", roomID)
	roomPresenceKey := fmt.Sprintf("%s:%d", roomOnlineKey, userID)
	if err := pm.cache.Set(ctx, roomPresenceKey, userID, presenceTTL); err != nil {
		return fmt.Errorf("failed to set room presence: %w", err)
	}

	return nil
}

func (pm *PresenceManager) SetUserOffline(ctx context.Context, userID uint, roomID string) error {
	userKey := cache.UserPresenceKey(userID)
	pm.cache.Delete(ctx, userKey)

	roomOnlineKey := fmt.Sprintf("chat:presence:room:%s", roomID)
	roomPresenceKey := fmt.Sprintf("%s:%d", roomOnlineKey, userID)
	pm.cache.Delete(ctx, roomPresenceKey)

	return nil
}

func (pm *PresenceManager) GetUserStatus(ctx context.Context, userID uint) (PresenceStatus, error) {
	userKey := cache.UserPresenceKey(userID)
	var status PresenceStatus
	err := pm.cache.Get(ctx, userKey, &status)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return PresenceStatusOffline, nil
		}
		return PresenceStatusOffline, err
	}
	return status, nil
}

func (pm *PresenceManager) GetRoomOnlineUsers(ctx context.Context, roomID string) ([]uint, error) {
	// TODO: Implement proper Redis set operations when cache interface is extended
	// For now, return empty list. Future implementation should:
	// 1. Add SMembers method to Cache interface
	// 2. Query "chat:presence:room:{roomID}" set for all online users
	// 3. Parse user IDs from the set members
	return []uint{}, nil
}

func (pm *PresenceManager) RefreshPresence(ctx context.Context, userID uint) error {
	userKey := cache.UserPresenceKey(userID)
	if err := pm.cache.Set(ctx, userKey, PresenceStatusOnline, presenceTTL); err != nil {
		return fmt.Errorf("failed to refresh presence: %w", err)
	}
	return nil
}

func (pm *PresenceManager) CreatePresenceEvent(userID uint, status PresenceStatus, roomID string) WSMessage {
	return WSMessage{
		Type: "presence",
		Payload: PresenceEvent{
			UserID: userID,
			Status: status,
			RoomID: roomID,
		},
	}
}

func (pm *PresenceManager) BroadcastPresenceUpdate(hub *Hub, userID uint, status PresenceStatus, roomID string) {
	event := pm.CreatePresenceEvent(userID, status, roomID)
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal presence event: %v", err)
		return
	}
	hub.BroadcastToRoom(roomID, data)
}

// Removed custom MarshalJSON - let Go's default JSON marshaling handle it properly

func BuildRoomID(conversationID *uint, groupID *uint, forumID *uint) string {
	if conversationID != nil {
		return fmt.Sprintf("conversation:%d", *conversationID)
	}
	if groupID != nil {
		return fmt.Sprintf("group:%d", *groupID)
	}
	if forumID != nil {
		return fmt.Sprintf("forum:%d", *forumID)
	}
	return ""
}

func ParseRoomID(roomID string) (conversationID *uint, groupID *uint, forumID *uint) {
	// Room format: "conversation:{id}", "group:{id}", or "forum:{id}"
	parts := strings.Split(roomID, ":")
	if len(parts) != 2 {
		return nil, nil, nil
	}

	id, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, nil, nil
	}

	idUint := uint(id)
	switch parts[0] {
	case "conversation":
		return &idUint, nil, nil
	case "group":
		return nil, &idUint, nil
	case "forum":
		return nil, nil, &idUint
	default:
		return nil, nil, nil
	}
}
