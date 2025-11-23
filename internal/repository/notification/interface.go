package notificationrepo

import (
	"context"

	"github.com/nanagoboiler/models"
)

type NotificationRepository interface {
	GetNotifications(ctx context.Context, uuid string) ([]models.Notification, error)
}
