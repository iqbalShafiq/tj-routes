package websocket

import (
	"context"
	"testing"
)

func TestNewPresenceManager(t *testing.T) {
	pm := NewPresenceManager(&MockCache{})
	if pm == nil {
		t.Fatal("Expected non-nil PresenceManager")
	}
	if pm.cache == nil {
		t.Fatal("Expected cache to be set")
	}
}

func TestPresenceManager_SetUserOnline(t *testing.T) {
	pm := NewPresenceManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)
	roomID := "conversation:123"

	err := pm.SetUserOnline(ctx, userID, roomID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestPresenceManager_SetUserOffline(t *testing.T) {
	pm := NewPresenceManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)
	roomID := "conversation:123"

	err := pm.SetUserOffline(ctx, userID, roomID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestPresenceManager_GetUserStatus(t *testing.T) {
	pm := NewPresenceManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)

	status, err := pm.GetUserStatus(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if status != PresenceStatusOffline {
		t.Errorf("Expected status %s, got %s", PresenceStatusOffline, status)
	}
}

func TestPresenceManager_RefreshPresence(t *testing.T) {
	pm := NewPresenceManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)

	err := pm.RefreshPresence(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestPresenceManager_CreatePresenceEvent(t *testing.T) {
	pm := NewPresenceManager(&MockCache{})
	userID := uint(1)
	status := PresenceStatusOnline
	roomID := "conversation:123"

	event := pm.CreatePresenceEvent(userID, status, roomID)

	if event.Payload == nil {
		t.Fatal("Expected payload to be non-nil")
	}

	payload, ok := event.Payload.(PresenceEvent)
	if !ok {
		t.Fatal("Expected payload to be PresenceEvent")
	}

	if payload.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, payload.UserID)
	}

	if payload.Status != status {
		t.Errorf("Expected Status %s, got %s", status, payload.Status)
	}

	if payload.RoomID != roomID {
		t.Errorf("Expected RoomID %s, got %s", roomID, payload.RoomID)
	}
}

func TestBuildRoomID(t *testing.T) {
	convID := uint(123)
	groupID := uint(456)
	forumID := uint(789)

	tests := []struct {
		name            string
		conversationID  *uint
		groupID         *uint
		forumID         *uint
		expectedPattern string
	}{
		{
			name:            "Conversation room",
			conversationID:  &convID,
			groupID:         nil,
			forumID:         nil,
			expectedPattern: "conversation:123",
		},
		{
			name:            "Group room",
			conversationID:  nil,
			groupID:         &groupID,
			forumID:         nil,
			expectedPattern: "group:456",
		},
		{
			name:            "Forum room",
			conversationID:  nil,
			groupID:         nil,
			forumID:         &forumID,
			expectedPattern: "forum:789",
		},
		{
			name:            "No room IDs",
			conversationID:  nil,
			groupID:         nil,
			forumID:         nil,
			expectedPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildRoomID(tt.conversationID, tt.groupID, tt.forumID)
			if result != tt.expectedPattern {
				t.Errorf("Expected %s, got %s", tt.expectedPattern, result)
			}
		})
	}
}

func TestWSMessage_MarshalJSON(t *testing.T) {
	msg := WSMessage{
		Type: "test",
		Payload: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := msg.MarshalJSON()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Expected non-empty JSON data")
	}
}
