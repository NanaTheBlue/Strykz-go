package mock

import (
	"context"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockSocialRepo struct {
	mock.Mock
}

func (m *MockSocialRepo) RemoveFriend(ctx context.Context, userID string, friendID string) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockSocialRepo) AddFriend(ctx context.Context, userID string, friendID string) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockSocialRepo) BlockUser(ctx context.Context, blocker string, blocked string) error {
	args := m.Called(ctx, blocker, blocked)
	return args.Error(0)
}

func (m *MockSocialRepo) CreateFriendRequest(ctx context.Context, friendreq models.FriendRequestInput) error {
	args := m.Called(ctx, friendreq)
	return args.Error(0)
}

func (m *MockSocialRepo) DeleteFriendRequest(ctx context.Context, senderID string, recipientID string) error {
	args := m.Called(ctx, senderID, recipientID)
	return args.Error(0)
}

func (m *MockSocialRepo) CreateParty(ctx context.Context, leaderID string) (string, error) {
	args := m.Called(ctx, leaderID)
	return args.String(0), args.Error(1)
}

func (m *MockSocialRepo) CheckPartyLeader(ctx context.Context, partyID string) (string, error) {
	args := m.Called(ctx, partyID)
	return args.String(0), args.Error(1)
}

func (m *MockSocialRepo) AddReport(ctx context.Context, reportreq models.ReportRequestInput) error {
	args := m.Called(ctx, reportreq)
	return args.Error(0)
}

func (m *MockSocialRepo) AddPartyMember(ctx context.Context, req models.PartyInviteRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockSocialRepo) IsBlocked(ctx context.Context, userID, otherID string) (bool, error) {
	args := m.Called(ctx, userID, otherID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepo) AddPartyInvite(ctx context.Context, req models.PartyInviteRequest) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepo) IsMutuallyBlocked(ctx context.Context, userA, userB string) (bool, error) {
	args := m.Called(ctx, userA, userB)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepo) IsFriends(ctx context.Context, friend1ID string, friend2ID string) (bool, error) {
	args := m.Called(ctx, friend1ID, friend2ID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSocialRepo) GetFriendRequest(ctx context.Context, friendReqID string) (*models.FriendRequestInput, error) {
	args := m.Called(ctx, friendReqID)
	req, _ := args.Get(0).(*models.FriendRequestInput)
	return req, args.Error(1)
}
