package social

import (
	"context"
)

type Service interface {
	BlockUser(ctx context.Context, blocker string, blocked string) error
	AcceptFriendRequest(ctx context.Context, userID string, friendID string) error
}
