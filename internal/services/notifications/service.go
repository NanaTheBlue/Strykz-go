package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/models"
)

type notificationsService struct {
	hub   *Hub
	store redis.Store
}

func NewnotificationsService(hub *Hub, store redis.Store) Service {
	return &notificationsService{
		hub:   hub,
		store: store,
	}
}

func (s *notificationsService) StartBackgroundListener(ctx context.Context) {
	go func() {
		err := s.store.Subscribe(ctx, "notifications", func(payload string) {
			var notif models.Notification
			if err := json.Unmarshal([]byte(payload), &notif); err != nil {
				fmt.Println("Failed to unmarshal notification:", err)
				return
			}

			if notif.UserID == "" {
				fmt.Println("Missing userID in notification")
				return
			}

			fmt.Printf("Sending to %s: %s\n", notif.UserID, notif.Data)
			if err := s.hub.Send(notif.UserID, notif.Data); err != nil {
				fmt.Println("Failed to send notification:", err)
			}
		})

		if err != nil {
			fmt.Println("Redis subscription error:", err)
		}
	}()
}

func (s *notificationsService) AddConnection(userID string, conn *websocket.Conn) {
	s.hub.Add(userID, conn)
}

func (s *notificationsService) RemoveConnection(userID string) {
	s.hub.Remove(userID)
}

func (s *notificationsService) SendNotification(ctx context.Context, notif models.Notification) error {
	return s.store.Publish(ctx, "notifications", notif)
}
