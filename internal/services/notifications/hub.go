package notifications

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nanagoboiler/models"
)

type Hub struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) Add(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	h.connections[userID] = conn
	h.mu.Unlock()
}

func (h *Hub) Remove(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conn, ok := h.connections[userID]
	if ok {
		conn.Close()
		delete(h.connections, userID)
	}
}

func (h *Hub) Send(userID string, notif models.Notification) error {
	h.mu.RLock()
	conn, ok := h.connections[userID]
	h.mu.RUnlock()
	if !ok {
		return nil
	}
	return conn.WriteJSON(map[string]interface{}{
		"type":         notif.Type,
		"id":           notif.ID,
		"sender_id":    notif.SenderID,
		"recipient_id": notif.RecipientID,
		"data":         notif.Data,
		"status":       notif.Status,
		"created_at":   notif.CreatedAt,
	})
}

func (h *Hub) Broadcast(notif models.Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conn := range h.connections {
		_ = conn.WriteJSON(map[string]interface{}{
			"type":         notif.Type,
			"id":           notif.ID,
			"sender_id":    notif.SenderID,
			"recipient_id": notif.RecipientID,
			"data":         notif.Data,
			"status":       notif.Status,
			"created_at":   notif.CreatedAt,
		})
	}
}
