package matchmaking

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Notifier interface {
	CreateAndPublishNotification(ctx context.Context, notif models.Notification) error
}
