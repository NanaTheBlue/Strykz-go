package social

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Service interface {
	AcceptNotification(ctx context.Context, notif models.Notification) error
}
