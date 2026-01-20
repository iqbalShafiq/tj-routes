package websocket

import (
	"log"
)

const (
	MessageTypeSendMessage = "send_message"
	MessageTypeMarkRead    = "mark_read"
	MessageTypeTyping      = "typing"
	MessageTypeTypingStart = "typing_start"
	MessageTypeTypingStop  = "typing_stop"
	MessageTypePing        = "ping"
	MessageTypePong        = "pong"
)

type SendMessagePayload struct {
	ConversationID *uint  `json:"conversation_id,omitempty"`
	GroupID        *uint  `json:"group_id,omitempty"`
	ForumID        *uint  `json:"forum_id,omitempty"`
	Content        string `json:"content"`
	ReplyToID      *uint  `json:"reply_to_id,omitempty"`
}

type MarkReadPayload struct {
	ConversationID *uint `json:"conversation_id,omitempty"`
	GroupID        *uint `json:"group_id,omitempty"`
}

type TypingPayload struct {
	ConversationID *uint `json:"conversation_id,omitempty"`
	GroupID        *uint `json:"group_id,omitempty"`
	IsTyping       bool  `json:"is_typing"`
}

type TypingStartPayload struct {
	ConversationID *uint `json:"conversation_id,omitempty"`
	GroupID        *uint `json:"group_id,omitempty"`
}

type TypingStopPayload struct {
	ConversationID *uint `json:"conversation_id,omitempty"`
	GroupID        *uint `json:"group_id,omitempty"`
}

func ValidateMessage(msg WSMessage) bool {
	switch msg.Type {
	case MessageTypeSendMessage:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return false
		}
		_, hasConvID := payload["conversation_id"]
		_, hasGroupID := payload["group_id"]
		_, hasForumID := payload["forum_id"]
		if !hasConvID && !hasGroupID && !hasForumID {
			return false
		}
		if _, exists := payload["content"]; !exists {
			return false
		}
		return true

	case MessageTypeMarkRead:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return false
		}
		_, hasConvID := payload["conversation_id"]
		_, hasGroupID := payload["group_id"]
		if !hasConvID && !hasGroupID {
			return false
		}
		return true

	case MessageTypeTyping:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return false
		}
		if _, exists := payload["is_typing"]; !exists {
			return false
		}
		return true

	case MessageTypeTypingStart:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return false
		}
		_, hasConvID := payload["conversation_id"]
		_, hasGroupID := payload["group_id"]
		if !hasConvID && !hasGroupID {
			return false
		}
		return true

	case MessageTypeTypingStop:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return false
		}
		_, hasConvID := payload["conversation_id"]
		_, hasGroupID := payload["group_id"]
		if !hasConvID && !hasGroupID {
			return false
		}
		return true

	case MessageTypePing, MessageTypePong:
		return true

	default:
		log.Printf("Unknown message type: %s", msg.Type)
		return false
	}
}
