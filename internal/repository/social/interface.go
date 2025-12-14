package socialrepo

import (
	"context"

	"github.com/nanagoboiler/models"
)

type SocialRepository interface {
	IsFriends(ctx context.Context, userID, user2ID string) (bool, error)
	RemoveFriend(ctx context.Context, userID string, friendID string) error
	AddFriend(ctx context.Context, userID string, friendID string) error
	BlockUser(ctx context.Context, blockreq models.BlockRequest) error
	IsBlocked(ctx context.Context, userID string, blockedID string) (bool, error)
	CreateFriendRequest(ctx context.Context, senderID, recipientID string) error
	DeleteFriendRequest(ctx context.Context, senderID string, recipientID string) error
}
