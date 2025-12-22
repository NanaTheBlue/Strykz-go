package notifications

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/nanagoboiler/models"
)

type Service interface {
	StartBackgroundListener(ctx context.Context)
	AddConnection(userID string, conn *websocket.Conn)
	RemoveConnection(userID string)
	CreateAndPublishNotification(ctx context.Context, notif models.Notification) error
	PublishNotification(ctx context.Context, notif models.Notification) error
	GetNotifications(ctx context.Context, userID string) ([]models.Notification, error)
	DeleteNotification(ctx context.Context, notifID string) error
}
