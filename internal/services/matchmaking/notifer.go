package matchmaking

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Notifier interface {
	CreateNoPublishNotification(ctx context.Context, notif models.Notification) (string, error)
}
