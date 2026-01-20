package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"tj-routes/internal/cache"
	"tj-routes/internal/models"
	"tj-routes/internal/service"
)

type Hub struct {
	clients             map[*Client]bool
	broadcast           chan []byte
	register            chan *Client
	unregister          chan *Client
	rooms               map[string]map[*Client]bool
	mu                  sync.RWMutex // Protects clients and rooms maps
	presenceMgr         *PresenceManager
	typingMgr           *TypingManager
	messageService      service.MessageService
	conversationService service.ConversationService
	forumMessageService service.ForumMessageService
	ctx                 context.Context
}

func NewHub(ctx context.Context, cacheClient cache.Cache) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		rooms:       make(map[string]map[*Client]bool),
		presenceMgr: NewPresenceManager(cacheClient),
		typingMgr:   NewTypingManager(cacheClient),
		ctx:         ctx,
	}
}

// SetServices sets the message, conversation, and forum message services for the hub
func (h *Hub) SetServices(messageService service.MessageService, conversationService service.ConversationService, forumMessageService service.ForumMessageService) {
	h.messageService = messageService
	h.conversationService = conversationService
	h.forumMessageService = forumMessageService
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %d", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				for roomID := range h.rooms {
					if _, exists := h.rooms[roomID][client]; exists {
						delete(h.rooms[roomID], client)
						if len(h.rooms[roomID]) == 0 {
							delete(h.rooms, roomID)
						}
					}
				}
				close(client.send)
				log.Printf("Client unregistered: %d", client.UserID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mu.Lock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	h.mu.Unlock()

	log.Printf("User %d joined room %s", client.UserID, roomID)

	if h.presenceMgr != nil {
		if err := h.presenceMgr.SetUserOnline(h.ctx, client.UserID, roomID); err != nil {
			log.Printf("Failed to set user presence: %v", err)
		} else {
			h.presenceMgr.BroadcastPresenceUpdate(h, client.UserID, PresenceStatusOnline, roomID)
		}
	}
}

func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mu.Lock()
	if h.rooms[roomID] != nil {
		delete(h.rooms[roomID], client)
		if len(h.rooms[roomID]) == 0 {
			delete(h.rooms, roomID)
		}
	}
	h.mu.Unlock()

	log.Printf("User %d left room %s", client.UserID, roomID)

	if h.presenceMgr != nil {
		if err := h.presenceMgr.SetUserOffline(h.ctx, client.UserID, roomID); err != nil {
			log.Printf("Failed to set user offline: %v", err)
		} else {
			h.presenceMgr.BroadcastPresenceUpdate(h, client.UserID, PresenceStatusOffline, roomID)
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.mu.RLock()
	clients := make([]*Client, 0)
	if h.rooms[roomID] != nil {
		for client := range h.rooms[roomID] {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			// Client channel full, will be handled by hub's main loop
			log.Printf("Failed to send message to user %d in room %s", client.UserID, roomID)
		}
	}
}

func (h *Hub) GetRoomClients(roomID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if h.rooms[roomID] != nil {
		for client := range h.rooms[roomID] {
			clients = append(clients, client)
		}
	}
	return clients
}

func (h *Hub) HandleTypingEvent(userID uint, conversationID *uint, groupID *uint, isTyping bool) {
	if isTyping {
		if h.typingMgr != nil {
			if err := h.typingMgr.UserStartedTyping(h.ctx, userID, conversationID, groupID); err != nil {
				log.Printf("Failed to set typing: %v", err)
			}
		}
	} else {
		if h.typingMgr != nil {
			if err := h.typingMgr.UserStoppedTyping(h.ctx, userID, conversationID, groupID); err != nil {
				log.Printf("Failed to stop typing: %v", err)
			}
		}
	}

	if h.typingMgr != nil {
		h.typingMgr.BroadcastTypingEvent(h, userID, conversationID, groupID, isTyping)
	}
}

// HandleSendMessage processes a message send request via WebSocket
func (h *Hub) HandleSendMessage(userID uint, content string, conversationID *uint, groupID *uint) (*models.Message, error) {
	if h.messageService == nil {
		return nil, errors.New("message service not initialized")
	}

	// Create message via service
	message, err := h.messageService.CreateMessage(content, userID, conversationID, groupID)
	if err != nil {
		return nil, err
	}

	// Determine room ID for broadcasting
	var roomID string
	if conversationID != nil {
		roomID = fmt.Sprintf("conversation:%d", *conversationID)
	} else if groupID != nil {
		roomID = fmt.Sprintf("group:%d", *groupID)
	}

	// Broadcast to room
	if roomID != "" {
		msgData, err := json.Marshal(map[string]interface{}{
			"type": "new_message",
			"data": message,
		})
		if err == nil {
			h.BroadcastToRoom(roomID, msgData)
		}
	}

	return message, nil
}

// HandleMarkRead processes a mark-as-read request via WebSocket
func (h *Hub) HandleMarkRead(userID uint, conversationID uint) error {
	if h.conversationService == nil {
		return errors.New("conversation service not initialized")
	}

	return h.conversationService.MarkAsRead(conversationID, userID)
}

// HandleSendForumMessage processes a forum message send request via WebSocket
func (h *Hub) HandleSendForumMessage(userID uint, forumID uint, content string) (*models.ForumMessage, error) {
	if h.forumMessageService == nil {
		return nil, errors.New("forum message service not initialized")
	}

	// Create message via service
	message, err := h.forumMessageService.CreateMessage(forumID, userID, content)
	if err != nil {
		return nil, err
	}

	// Broadcast to forum room
	roomID := fmt.Sprintf("forum:%d", forumID)
	msgData, err := json.Marshal(map[string]interface{}{
		"type": "new_forum_message",
		"data": message,
	})
	if err == nil {
		h.BroadcastToRoom(roomID, msgData)
	}

	return message, nil
}
