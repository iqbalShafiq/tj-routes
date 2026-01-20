package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize is now configured via WebSocketConfig (default: 65536 bytes = 64KB)
)

type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	UserID         uint
	maxMessageSize int64
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uint, maxMessageSize int64) *Client {
	return &Client{
		hub:            hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		UserID:         userID,
		maxMessageSize: maxMessageSize,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		c.handleMessage(wsMsg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg WSMessage) {
	switch msg.Type {
	case "ping":
		c.sendJSON(WSMessage{Type: "pong", Payload: nil})
	case "join_room":
		c.handleJoinRoom(msg.Payload)
	case "leave_room":
		c.handleLeaveRoom(msg.Payload)
	case "send_message":
		c.handleSendMessage(msg.Payload)
	case "mark_read":
		c.handleMarkRead(msg.Payload)
	case "typing":
		c.handleTyping(msg.Payload)
	case "typing_start":
		c.handleTypingStart(msg.Payload)
	case "typing_stop":
		c.handleTypingStop(msg.Payload)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (c *Client) handleSendMessage(payload interface{}) {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid send_message payload from user %d", c.UserID)
		return
	}

	content, ok := payloadMap["content"].(string)
	if !ok || content == "" {
		log.Printf("Missing or invalid content in send_message from user %d", c.UserID)
		return
	}

	var conversationID *uint
	var groupID *uint
	var forumID *uint

	if convID, exists := payloadMap["conversation_id"]; exists && convID != nil {
		if val, ok := convID.(float64); ok {
			convIDUint := uint(val)
			conversationID = &convIDUint
		}
	}

	if grpID, exists := payloadMap["group_id"]; exists && grpID != nil {
		if val, ok := grpID.(float64); ok {
			grpIDUint := uint(val)
			groupID = &grpIDUint
		}
	}

	if frmID, exists := payloadMap["forum_id"]; exists && frmID != nil {
		if val, ok := frmID.(float64); ok {
			frmIDUint := uint(val)
			forumID = &frmIDUint
		}
	}

	// Route to appropriate handler based on message type
	if forumID != nil {
		// Handle forum message
		message, err := c.hub.HandleSendForumMessage(c.UserID, *forumID, content)
		if err != nil {
			log.Printf("Failed to send forum message for user %d: %v", c.UserID, err)
			c.sendJSON(WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Failed to send forum message",
					"error":   err.Error(),
				},
			})
			return
		}

		// Send confirmation to sender
		c.sendJSON(WSMessage{
			Type:    "forum_message_sent",
			Payload: message,
		})
		return
	}

	// Handle conversation/group message
	message, err := c.hub.HandleSendMessage(c.UserID, content, conversationID, groupID)
	if err != nil {
		log.Printf("Failed to send message for user %d: %v", c.UserID, err)
		c.sendJSON(WSMessage{
			Type: "error",
			Payload: map[string]interface{}{
				"message": "Failed to send message",
				"error":   err.Error(),
			},
		})
		return
	}

	// Send confirmation to sender
	c.sendJSON(WSMessage{
		Type:    "message_sent",
		Payload: message,
	})
}

func (c *Client) handleMarkRead(payload interface{}) {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid mark_read payload from user %d", c.UserID)
		return
	}

	conversationID, ok := payloadMap["conversation_id"].(float64)
	if !ok {
		log.Printf("Missing or invalid conversation_id in mark_read from user %d", c.UserID)
		return
	}

	err := c.hub.HandleMarkRead(c.UserID, uint(conversationID))
	if err != nil {
		log.Printf("Failed to mark as read for user %d: %v", c.UserID, err)
		c.sendJSON(WSMessage{
			Type: "error",
			Payload: map[string]interface{}{
				"message": "Failed to mark as read",
				"error":   err.Error(),
			},
		})
		return
	}

	// Send confirmation
	c.sendJSON(WSMessage{
		Type: "marked_read",
		Payload: map[string]interface{}{
			"conversation_id": uint(conversationID),
		},
	})
}

func (c *Client) handleTyping(payload interface{}) {
	log.Printf("User %d is typing: %v", c.UserID, payload)
}

func (c *Client) handleTypingStart(payload interface{}) {
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		var conversationID *uint
		var groupID *uint

		if convID, exists := payloadMap["conversation_id"]; exists && convID != nil {
			if val, ok := convID.(float64); ok {
				convIDUint := uint(val)
				conversationID = &convIDUint
			}
		}

		if grpID, exists := payloadMap["group_id"]; exists && grpID != nil {
			if val, ok := grpID.(float64); ok {
				grpIDUint := uint(val)
				groupID = &grpIDUint
			}
		}

		c.hub.HandleTypingEvent(c.UserID, conversationID, groupID, true)
	}
}

func (c *Client) handleTypingStop(payload interface{}) {
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		var conversationID *uint
		var groupID *uint

		if convID, exists := payloadMap["conversation_id"]; exists && convID != nil {
			if val, ok := convID.(float64); ok {
				convIDUint := uint(val)
				conversationID = &convIDUint
			}
		}

		if grpID, exists := payloadMap["group_id"]; exists && grpID != nil {
			if val, ok := grpID.(float64); ok {
				grpIDUint := uint(val)
				groupID = &grpIDUint
			}
		}

		c.hub.HandleTypingEvent(c.UserID, conversationID, groupID, false)
	}
}

func (c *Client) handleJoinRoom(payload interface{}) {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid join_room payload from user %d", c.UserID)
		return
	}

	roomID, ok := payloadMap["room_id"].(string)
	if !ok || roomID == "" {
		log.Printf("Missing or invalid room_id in join_room from user %d", c.UserID)
		return
	}

	// Join the room
	c.hub.JoinRoom(c, roomID)

	// Send confirmation
	c.sendJSON(WSMessage{
		Type: "room_joined",
		Payload: map[string]interface{}{
			"room_id": roomID,
		},
	})
}

func (c *Client) handleLeaveRoom(payload interface{}) {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid leave_room payload from user %d", c.UserID)
		return
	}

	roomID, ok := payloadMap["room_id"].(string)
	if !ok || roomID == "" {
		log.Printf("Missing or invalid room_id in leave_room from user %d", c.UserID)
		return
	}

	// Leave the room
	c.hub.LeaveRoom(c, roomID)

	// Send confirmation
	c.sendJSON(WSMessage{
		Type: "room_left",
		Payload: map[string]interface{}{
			"room_id": roomID,
		},
	})
}

func (c *Client) sendJSON(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		log.Printf("Send channel full for user %d", c.UserID)
	}
}
