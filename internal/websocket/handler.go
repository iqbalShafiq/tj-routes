package websocket

import (
	"log"
	"net/http"

	"tj-routes/internal/config"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	hub *Hub
	cfg *config.Config
}

func NewWebSocketHandler(hub *Hub, cfg *config.Config) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		cfg: cfg,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	token := ExtractTokenFromRequest(c.Request)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
		return
	}

	userID, err := ValidateWebSocketToken(token, h.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	maxMessageSize := h.cfg.WebSocket.MaxMessageSize
	if maxMessageSize == 0 {
		maxMessageSize = 65536 // Default to 64KB
	}

	client := NewClient(h.hub, conn, userID, maxMessageSize)

	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
