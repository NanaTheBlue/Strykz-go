//go:build unit

package social

import (
	"context"
	"errors"
	"testing"

	"github.com/gorilla/websocket"
	socialmock "github.com/nanagoboiler/internal/repository/social/mock"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) StartBackgroundListener(ctx context.Context)                    {}
func (m *MockNotificationService) AddConnection(userID string, conn *websocket.Conn)              {}
func (m *MockNotificationService) RemoveConnection(userID string)                                 {}
func (m *MockNotificationService) CreateAndPublishNotification(ctx context.Context, notif models.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}
func (m *MockNotificationService) CreateNoPublishNotification(ctx context.Context, notif models.Notification) (string, error) {
	args := m.Called(ctx, notif)
	return args.String(0), args.Error(1)
}
func (m *MockNotificationService) PublishNotification(ctx context.Context, notif models.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}
func (m *MockNotificationService) GetNotifications(ctx context.Context, userID string) ([]models.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Notification), args.Error(1)
}
func (m *MockNotificationService) DeleteNotification(ctx context.Context, notifID string) error {
	args := m.Called(ctx, notifID)
	return args.Error(0)
}

func TestBlockUser_Success(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)

	svc := &socialService{
		socialrepo: mockSocialRepo,
	}

	mockSocialRepo.On("BlockUser", ctx, "user1", "user2").Return(nil)

	err := svc.BlockUser(ctx, "user1", "user2")
	assert.NoError(t, err)
	mockSocialRepo.AssertExpectations(t)
}

func TestBlockUser_CannotBlockSelf(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)

	svc := &socialService{
		socialrepo: mockSocialRepo,
	}

	err := svc.BlockUser(ctx, "user1", "user1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot block yourself")
	mockSocialRepo.AssertNotCalled(t, "BlockUser")
}

func TestBlockUser_RepoError(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)

	svc := &socialService{
		socialrepo: mockSocialRepo,
	}

	mockSocialRepo.On("BlockUser", ctx, "user1", "user2").Return(errors.New("db error"))

	err := svc.BlockUser(ctx, "user1", "user2")
	assert.Error(t, err)
}

func TestNormalizePair(t *testing.T) {
	tests := []struct {
		name              string
		a, b              string
		wantA, wantB string
	}{
		{"already ordered", "alice", "bob", "alice", "bob"},
		{"needs swap", "bob", "alice", "alice", "bob"},
		{"equal", "same", "same", "same", "same"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := normalizePair(tt.a, tt.b)
			assert.Equal(t, tt.wantA, gotA)
			assert.Equal(t, tt.wantB, gotB)
		})
	}
}

func TestReportUser_Success(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)

	svc := &socialService{
		socialrepo: mockSocialRepo,
	}

	input := models.ReportRequestInput{
		ReporterID: "user1",
		ReporteeID: "user2",
		Type:       "cheating",
		Reason:     "aimbot",
	}
	mockSocialRepo.On("AddReport", ctx, input).Return(nil)

	err := svc.ReportUser(ctx, input)
	assert.NoError(t, err)
	mockSocialRepo.AssertExpectations(t)
}

func TestReportUser_RepoError(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)

	svc := &socialService{
		socialrepo: mockSocialRepo,
	}

	input := models.ReportRequestInput{
		ReporterID: "user1",
		ReporteeID: "user2",
		Type:       "cheating",
		Reason:     "aimbot",
	}
	mockSocialRepo.On("AddReport", ctx, input).Return(errors.New("db error"))

	err := svc.ReportUser(ctx, input)
	assert.Error(t, err)
}

func TestSendFriendRequest_Success(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)
	mockNotifService := new(MockNotificationService)

	svc := &socialService{
		socialrepo:          mockSocialRepo,
		notificationservice: mockNotifService,
	}

	input := models.FriendRequestInput{
		SenderID:    "user1",
		RecipientID: "user2",
	}
	mockSocialRepo.On("CreateFriendRequest", ctx, input).Return(nil)
	mockNotifService.On("PublishNotification", mock.Anything, mock.MatchedBy(func(n models.Notification) bool {
		return n.SenderID == "user1" && n.RecipientID == "user2" && n.Type == models.FriendRequest
	})).Return(nil)

	err := svc.SendFriendRequest(ctx, input)
	assert.NoError(t, err)
	mockSocialRepo.AssertExpectations(t)
	mockNotifService.AssertExpectations(t)
}

func TestSendFriendRequest_RepoError(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)
	mockNotifService := new(MockNotificationService)

	svc := &socialService{
		socialrepo:          mockSocialRepo,
		notificationservice: mockNotifService,
	}

	input := models.FriendRequestInput{
		SenderID:    "user1",
		RecipientID: "user2",
	}
	mockSocialRepo.On("CreateFriendRequest", ctx, input).Return(errors.New("already friends"))

	err := svc.SendFriendRequest(ctx, input)
	assert.Error(t, err)
	mockNotifService.AssertNotCalled(t, "PublishNotification")
}

func TestSendFriendRequest_PublishError(t *testing.T) {
	ctx := context.Background()
	mockSocialRepo := new(socialmock.MockSocialRepo)
	mockNotifService := new(MockNotificationService)

	svc := &socialService{
		socialrepo:          mockSocialRepo,
		notificationservice: mockNotifService,
	}

	input := models.FriendRequestInput{
		SenderID:    "user1",
		RecipientID: "user2",
	}
	mockSocialRepo.On("CreateFriendRequest", ctx, input).Return(nil)
	mockNotifService.On("PublishNotification", mock.Anything, mock.Anything).
		Return(errors.New("redis error"))

	err := svc.SendFriendRequest(ctx, input)
	assert.Error(t, err)
}

func TestRejectNotification_Success(t *testing.T) {
	ctx := context.Background()
	mockNotifService := new(MockNotificationService)

	svc := &socialService{
		notificationservice: mockNotifService,
	}

	mockNotifService.On("DeleteNotification", ctx, "notif-1").Return(nil)

	err := svc.RejectNotification(ctx, "notif-1")
	assert.NoError(t, err)
}

func TestRejectNotification_Error(t *testing.T) {
	ctx := context.Background()
	mockNotifService := new(MockNotificationService)

	svc := &socialService{
		notificationservice: mockNotifService,
	}

	mockNotifService.On("DeleteNotification", ctx, "notif-1").Return(errors.New("not found"))

	err := svc.RejectNotification(ctx, "notif-1")
	assert.Error(t, err)
}
