package mock

import (
	"context"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockSocialRepository struct {
	mock.Mock
}

func (m *MockSocialRepository) RemoveFriend(ctx context.Context, userID, friendID string) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockSocialRepository) AddFriend(ctx context.Context, userID, friendID string) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockSocialRepository) BlockUser(ctx context.Context, blocker, blocked string) error {
	args := m.Called(ctx, blocker, blocked)
	return args.Error(0)
}

func (m *MockSocialRepository) CreateFriendRequest(ctx context.Context, friendreq models.FriendRequestInput) error {
	args := m.Called(ctx, friendreq)
	return args.Error(0)
}

func (m *MockSocialRepository) DeleteFriendRequest(ctx context.Context, senderID, recipientID string) error {
	args := m.Called(ctx, senderID, recipientID)
	return args.Error(0)
}

func (m *MockSocialRepository) CreateParty(ctx context.Context, leaderID string) (string, error) {
	args := m.Called(ctx, leaderID)
	return args.String(0), args.Error(1)
}

func (m *MockSocialRepository) CheckPartyLeader(ctx context.Context, partyID string) (string, error) {
	args := m.Called(ctx, partyID)
	return args.String(0), args.Error(1)
}

func (m *MockSocialRepository) AddReport(ctx context.Context, reportreq models.ReportRequestInput) error {
	args := m.Called(ctx, reportreq)
	return args.Error(0)
}

func (m *MockSocialRepository) AddPartyMember(ctx context.Context, req models.PartyInviteRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockSocialRepository) IsBlocked(ctx context.Context, userID, otherID string) (bool, error) {
	args := m.Called(ctx, userID, otherID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepository) AddPartyInvite(ctx context.Context, req models.PartyInviteRequest) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepository) IsMutuallyBlocked(ctx context.Context, userA, userB string) (bool, error) {
	args := m.Called(ctx, userA, userB)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepository) IsFriends(ctx context.Context, friend1ID, friend2ID string) (bool, error) {
	args := m.Called(ctx, friend1ID, friend2ID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepository) GetFriendRequest(ctx context.Context, friendReqID string) (*models.FriendRequestInput, error) {
	args := m.Called(ctx, friendReqID)
	var friendReq *models.FriendRequestInput
	if f := args.Get(0); f != nil {
		friendReq = f.(*models.FriendRequestInput)
	}
	return friendReq, args.Error(1)
}
