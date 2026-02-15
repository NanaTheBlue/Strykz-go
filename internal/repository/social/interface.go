package socialrepo

import (
	"context"

	"github.com/nanagoboiler/models"
)

type SocialRepository interface {
	RemoveFriend(ctx context.Context, userID string, friendID string) error
	AddFriend(ctx context.Context, userID string, friendID string) error
	BlockUser(ctx context.Context, blockreq models.BlockRequest) error
	CreateFriendRequest(ctx context.Context, friendreq models.FriendRequestInput) error
	DeleteFriendRequest(ctx context.Context, senderID string, recipientID string) error
	CreateParty(ctx context.Context, leaderID string) (string, error)
	CheckPartyLeader(ctx context.Context, partyID string) (string, error)
	AddReport(ctx context.Context, reportreq models.ReportRequestInput) error
	AddPartyMember(ctx context.Context, req models.PartyInviteRequest) error
	IsBlocked(ctx context.Context, userID, otherID string) (bool, error)
	AddPartyInvite(ctx context.Context, req models.PartyInviteRequest) (bool, error)
	IsMutuallyBlocked(ctx context.Context, userA, userB string) (bool, error)
	IsFriends(ctx context.Context, friend1ID string, friend2ID string) (bool, error)
}
