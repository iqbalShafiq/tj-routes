package websocket

import (
	"context"
	"testing"
	"time"
)

func TestNewTypingManager(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	if tm == nil {
		t.Fatal("Expected non-nil TypingManager")
	}
	if tm.cache == nil {
		t.Fatal("Expected cache to be set")
	}
	if tm.typingData == nil {
		t.Fatal("Expected typingData to be initialized")
	}
}

func TestTypingManager_UserStartedTyping(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)
	convID := uint(123)

	err := tm.UserStartedTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	isTyping, err := tm.IsUserTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !isTyping {
		t.Error("Expected user to be typing")
	}
}

func TestTypingManager_UserStoppedTyping(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)
	convID := uint(123)

	tm.UserStartedTyping(ctx, userID, &convID, nil)

	err := tm.UserStoppedTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	isTyping, err := tm.IsUserTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if isTyping {
		t.Error("Expected user to not be typing")
	}
}

func TestTypingManager_GetTypingUsers(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	ctx := context.Background()

	convID := uint(123)

	userID1 := uint(1)
	userID2 := uint(2)

	tm.UserStartedTyping(ctx, userID1, &convID, nil)
	tm.UserStartedTyping(ctx, userID2, &convID, nil)

	typingUsers, err := tm.GetTypingUsers(ctx, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(typingUsers) != 2 {
		t.Errorf("Expected 2 typing users, got %d", len(typingUsers))
	}
}

func TestTypingManager_IsUserTyping(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)
	convID := uint(123)

	isTyping, err := tm.IsUserTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if isTyping {
		t.Error("Expected user to not be typing initially")
	}

	tm.UserStartedTyping(ctx, userID, &convID, nil)

	isTyping, err = tm.IsUserTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !isTyping {
		t.Error("Expected user to be typing after starting")
	}
}

func TestTypingManager_GroupTyping(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)
	groupID := uint(456)

	err := tm.UserStartedTyping(ctx, userID, nil, &groupID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	isTyping, err := tm.IsUserTyping(ctx, userID, nil, &groupID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !isTyping {
		t.Error("Expected user to be typing in group")
	}
}

func TestTypingManager_NoRoomIDError(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)

	err := tm.UserStartedTyping(ctx, userID, nil, nil)
	if err == nil {
		t.Error("Expected error when no room ID provided")
	}

	err = tm.UserStoppedTyping(ctx, userID, nil, nil)
	if err == nil {
		t.Error("Expected error when no room ID provided")
	}
}

func TestTypingManager_CreateTypingEvent(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	userID := uint(1)
	convID := uint(123)
	isTyping := true

	event := tm.CreateTypingEvent(userID, &convID, nil, isTyping)

	if event.Payload == nil {
		t.Fatal("Expected payload to be non-nil")
	}

	payload, ok := event.Payload.(TypingEvent)
	if !ok {
		t.Fatal("Expected payload to be TypingEvent")
	}

	if payload.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, payload.UserID)
	}

	if *payload.ConversationID != convID {
		t.Errorf("Expected ConversationID %d, got %d", convID, *payload.ConversationID)
	}

	if payload.IsTyping != isTyping {
		t.Errorf("Expected IsTyping %t, got %t", isTyping, payload.IsTyping)
	}
}

func TestTypingDebouncer(t *testing.T) {
	debouncer := &TypingDebouncer{}

	if !debouncer.ShouldBroadcast() {
		t.Error("Expected first broadcast to be allowed")
	}

	if debouncer.ShouldBroadcast() {
		t.Error("Expected second immediate broadcast to be debounced")
	}

	time.Sleep(600 * time.Millisecond)

	if !debouncer.ShouldBroadcast() {
		t.Error("Expected broadcast to be allowed after debounce period")
	}
}

func TestTypingManager_ExpiredTypingCleanup(t *testing.T) {
	tm := NewTypingManager(&MockCache{})
	ctx := context.Background()

	userID := uint(1)
	convID := uint(123)

	tm.UserStartedTyping(ctx, userID, &convID, nil)

	isTyping, _ := tm.IsUserTyping(ctx, userID, &convID, nil)
	if !isTyping {
		t.Error("Expected user to be typing initially")
	}

	time.Sleep(6 * time.Second)

	isTyping, err := tm.IsUserTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if isTyping {
		t.Error("Expected user typing to expire after timeout")
	}
}
