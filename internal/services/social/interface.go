package social

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Service interface {
	AcceptNotification(ctx context.Context, notif models.Notification) error
	BlockUser(ctx context.Context, blocker string, blocked string) error
	AcceptFriendRequest(ctx context.Context, userID string, friendID string) error
}
