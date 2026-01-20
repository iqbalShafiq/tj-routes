package websocket

import (
	"log"
	"net/http"
	"strings"

	"tj-routes/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader

// InitUpgrader initializes the WebSocket upgrader with config
func InitUpgrader(cfg *config.Config) {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  cfg.WebSocket.ReadBufferSize,
		WriteBufferSize: cfg.WebSocket.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			allowedOrigins := cfg.Server.AllowedOrigin

			// Allow all origins if configured with "*"
			if allowedOrigins == "*" {
				return true
			}

			// Check if origin is in allowed list (comma-separated)
			if allowedOrigins != "" {
				origins := strings.Split(allowedOrigins, ",")
				for _, allowed := range origins {
					allowed = strings.TrimSpace(allowed)
					if origin == allowed {
						return true
					}
				}
			}

			log.Printf("WebSocket connection from unauthorized origin: %s", origin)
			return false
		},
	}
}

func UpgradeConnection(c *gin.Context) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return nil, err
	}
	return conn, nil
}
