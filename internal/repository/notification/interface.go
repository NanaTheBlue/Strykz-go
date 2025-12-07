package notificationrepo

import (
	"context"

	"github.com/nanagoboiler/models"
)

type NotificationRepository interface {
	GetNotifications(ctx context.Context, uuid string) ([]models.Notification, error)
	GetNotification(ctx context.Context, notificationID string) (models.Notification, error)
	SendNotification(ctx context.Context, notif models.Notification) error
	DeleteNotification(ctx context.Context, notificationID string) error
}
