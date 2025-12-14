package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Implement proper origin checking
	},
}

// WebSocketHandler handles WebSocket connections for real-time updates
type WebSocketHandler struct {
	clients map[*websocket.Conn]bool
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients: make(map[*websocket.Conn]bool),
	}
}

// HandleConnection handles new WebSocket connections
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	h.clients[conn] = true
	defer delete(h.clients, conn)

	log.Info().Msg("WebSocket client connected")

	// Keep connection alive and handle messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msg("WebSocket read error")
			break
		}

		// Echo message back (stub implementation)
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Error().Err(err).Msg("WebSocket write error")
			break
		}
	}

	log.Info().Msg("WebSocket client disconnected")
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHandler) Broadcast(message []byte) {
	for client := range h.clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Error().Err(err).Msg("Failed to broadcast message")
			client.Close()
			delete(h.clients, client)
		}
	}
}
