package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	notificationrepo "github.com/nanagoboiler/internal/repository/notification"
	"github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/models"
)

type notificationsService struct {
	hub              *Hub
	store            redis.Store
	notificationrepo notificationrepo.NotificationRepository
}

func NewnotificationsService(hub *Hub, store redis.Store, notificationrepo notificationrepo.NotificationRepository) Service {
	return &notificationsService{
		hub:              hub,
		store:            store,
		notificationrepo: notificationrepo,
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

			if notif.Recepient_id == "" {
				fmt.Println("Missing Recepient ID in notification")
				return
			}

			fmt.Printf("Sending to %s: %s\n", notif.Recepient_id, notif.Data)
			if err := s.hub.Send(notif.Recepient_id, notif.Data); err != nil {
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

func (s *notificationsService) GetNotifications(ctx context.Context, userID string) ([]models.Notification, error) {

	notifications, err := s.notificationrepo.GetNotifications(ctx, userID)
	if err != nil {
		return []models.Notification{}, err
	}

	return notifications, nil

}

func (s *notificationsService) SendFriendRequest(ctx context.Context, notif models.Notification) error {

	return nil
}

func (s *notificationsService) AcceptNotification(ctx context.Context, notif models.Notification) error {

	if notif.Type == "FriendRequest" {
		// add friends
		err := s.notificationrepo.AddFriend(ctx, notif)
		if err != nil {
			return err
		}
	} else if notif.Type == "PartyInvite" {

		// Join Party

	}
	return nil
}

func (s *notificationsService) RejectNotification(ctx context.Context, notif models.Notification) error {

	return nil
}
