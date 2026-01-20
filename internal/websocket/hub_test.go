package websocket

import (
	"context"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	ctx := context.Background()
	hub := NewHub(ctx, &MockCache{})

	if hub == nil {
		t.Fatal("Expected non-nil Hub")
	}

	if hub.clients == nil {
		t.Fatal("Expected clients map to be initialized")
	}

	if hub.rooms == nil {
		t.Fatal("Expected rooms map to be initialized")
	}

	if hub.presenceMgr == nil {
		t.Fatal("Expected presence manager to be initialized")
	}

	if hub.typingMgr == nil {
		t.Fatal("Expected typing manager to be initialized")
	}
}

func TestHub_JoinRoom(t *testing.T) {
	ctx := context.Background()
	hub := NewHub(ctx, &MockCache{})

	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		UserID: 1,
	}

	roomID := "conversation:123"

	hub.JoinRoom(client, roomID)

	if len(hub.rooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(hub.rooms))
	}

	if _, exists := hub.rooms[roomID]; !exists {
		t.Error("Expected room to exist")
	}

	if !hub.rooms[roomID][client] {
		t.Error("Expected client to be in room")
	}
}

func TestHub_LeaveRoom(t *testing.T) {
	ctx := context.Background()
	hub := NewHub(ctx, &MockCache{})

	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		UserID: 1,
	}

	roomID := "conversation:123"

	hub.JoinRoom(client, roomID)
	hub.LeaveRoom(client, roomID)

	if len(hub.rooms) != 0 {
		t.Errorf("Expected 0 rooms, got %d", len(hub.rooms))
	}
}

func TestHub_GetRoomClients(t *testing.T) {
	ctx := context.Background()
	hub := NewHub(ctx, &MockCache{})

	roomID := "conversation:123"

	client1 := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		UserID: 1,
	}

	client2 := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		UserID: 2,
	}

	hub.JoinRoom(client1, roomID)
	hub.JoinRoom(client2, roomID)

	clients := hub.GetRoomClients(roomID)

	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}
}

func TestHub_HandleTypingEvent(t *testing.T) {
	ctx := context.Background()
	hub := NewHub(ctx, &MockCache{})

	userID := uint(1)
	convID := uint(123)

	hub.HandleTypingEvent(userID, &convID, nil, true)

	isTyping, err := hub.typingMgr.IsUserTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !isTyping {
		t.Error("Expected user to be typing")
	}

	hub.HandleTypingEvent(userID, &convID, nil, false)

	isTyping, err = hub.typingMgr.IsUserTyping(ctx, userID, &convID, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if isTyping {
		t.Error("Expected user to not be typing")
	}
}

func TestHub_Run(t *testing.T) {
	ctx := context.Background()
	hub := NewHub(ctx, &MockCache{})

	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		UserID: 1,
	}

	done := make(chan bool)

	go func() {
		hub.Run()
		done <- true
	}()

	hub.register <- client

	select {
	case <-time.After(100 * time.Millisecond):
		// OK - test passed
	case <-done:
		// OK - hub stopped
	}
}
