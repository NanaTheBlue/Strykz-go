//go:build unit

package social

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	redisMocks "github.com/nanagoboiler/internal/repository/redis/mock"
	socialMocks "github.com/nanagoboiler/internal/repository/social/mock"
	notificationMocks "github.com/nanagoboiler/internal/services/notifications/mock"
	"github.com/nanagoboiler/models"
)

func setupSocialService() *socialService {
	notificationService := new(notificationMocks.MockNotificationService)
	socialRepo := new(socialMocks.MockSocialRepository)
	store := new(redisMocks.MockRedisStore)

	return &socialService{
		notificationservice: notificationService,
		socialrepo:          socialRepo,
		store:               store,
		pool:                nil, // Will be nil for unit tests
	}
}

func TestBlockUser_Success(t *testing.T) {
	svc := setupSocialService()

	blocker := "user123"
	blocked := "user456"

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("BlockUser", mock.Anything, blocker, blocked).
		Return(nil)

	err := svc.BlockUser(context.Background(), blocker, blocked)

	assert.NoError(t, err)
	svc.socialrepo.(*socialMocks.MockSocialRepository).AssertExpectations(t)
}

func TestBlockUser_CannotBlockSelf(t *testing.T) {
	svc := setupSocialService()

	userID := "user123"

	err := svc.BlockUser(context.Background(), userID, userID)

	assert.Error(t, err)
	assert.Equal(t, "cannot block yourself", err.Error())
}

func TestBlockUser_RepositoryError(t *testing.T) {
	svc := setupSocialService()

	blocker := "user123"
	blocked := "user456"

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("BlockUser", mock.Anything, blocker, blocked).
		Return(errors.New("database error"))

	err := svc.BlockUser(context.Background(), blocker, blocked)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
}

func TestSendFriendRequest_Success(t *testing.T) {
	svc := setupSocialService()

	friendReq := models.FriendRequestInput{
		SenderID:    "user123",
		RecipientID: "user456",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("CreateFriendRequest", mock.Anything, friendReq).
		Return(nil)

	svc.notificationservice.(*notificationMocks.MockNotificationService).
		On("PublishNotification", mock.Anything, mock.MatchedBy(func(n models.Notification) bool {
			return n.SenderID == friendReq.SenderID && n.RecipientID == friendReq.RecipientID && n.Type == models.FriendRequest
		})).
		Return(nil)

	err := svc.SendFriendRequest(context.Background(), friendReq)

	assert.NoError(t, err)
	svc.socialrepo.(*socialMocks.MockSocialRepository).AssertExpectations(t)
	svc.notificationservice.(*notificationMocks.MockNotificationService).AssertExpectations(t)
}

func TestSendFriendRequest_RepositoryError(t *testing.T) {
	svc := setupSocialService()

	friendReq := models.FriendRequestInput{
		SenderID:    "user123",
		RecipientID: "user456",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("CreateFriendRequest", mock.Anything, friendReq).
		Return(errors.New("friend request failed"))

	err := svc.SendFriendRequest(context.Background(), friendReq)

	assert.Error(t, err)
	assert.Equal(t, "friend request failed", err.Error())
}

func TestSendFriendRequest_PublishNotificationError(t *testing.T) {
	svc := setupSocialService()

	friendReq := models.FriendRequestInput{
		SenderID:    "user123",
		RecipientID: "user456",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("CreateFriendRequest", mock.Anything, friendReq).
		Return(nil)

	svc.notificationservice.(*notificationMocks.MockNotificationService).
		On("PublishNotification", mock.Anything, mock.Anything).
		Return(errors.New("publish failed"))

	err := svc.SendFriendRequest(context.Background(), friendReq)

	assert.Error(t, err)
	assert.Equal(t, "publish failed", err.Error())
}

func TestReportUser_Success(t *testing.T) {
	svc := setupSocialService()

	reportReq := models.ReportRequestInput{
		ReporterID: "user123",
		ReporteeID: "user456",
		Type:       "harassment",
		Reason:     "Inappropriate behavior",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("AddReport", mock.Anything, reportReq).
		Return(nil)

	err := svc.ReportUser(context.Background(), reportReq)

	assert.NoError(t, err)
	svc.socialrepo.(*socialMocks.MockSocialRepository).AssertExpectations(t)
}

func TestReportUser_Error(t *testing.T) {
	svc := setupSocialService()

	reportReq := models.ReportRequestInput{
		ReporterID: "user123",
		ReporteeID: "user456",
		Type:       "harassment",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("AddReport", mock.Anything, reportReq).
		Return(errors.New("report creation failed"))

	err := svc.ReportUser(context.Background(), reportReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add report")
}

func TestCreateParty_Success(t *testing.T) {
	svc := setupSocialService()

	userID := "user123"
	partyID := uuid.NewString()

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("CreateParty", mock.Anything, userID).
		Return(partyID, nil)

	result, err := svc.CreateParty(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, partyID, result)
	svc.socialrepo.(*socialMocks.MockSocialRepository).AssertExpectations(t)
}

func TestCreateParty_Error(t *testing.T) {
	svc := setupSocialService()

	userID := "user123"

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("CreateParty", mock.Anything, userID).
		Return("", errors.New("party creation failed"))

	result, err := svc.CreateParty(context.Background(), userID)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Equal(t, "party creation failed", err.Error())
}

func TestRejectNotification_Success(t *testing.T) {
	svc := setupSocialService()

	notifID := uuid.NewString()

	svc.notificationservice.(*notificationMocks.MockNotificationService).
		On("DeleteNotification", mock.Anything, notifID).
		Return(nil)

	err := svc.RejectNotification(context.Background(), notifID)

	assert.NoError(t, err)
	svc.notificationservice.(*notificationMocks.MockNotificationService).AssertExpectations(t)
}

func TestRejectNotification_Error(t *testing.T) {
	svc := setupSocialService()

	notifID := uuid.NewString()

	svc.notificationservice.(*notificationMocks.MockNotificationService).
		On("DeleteNotification", mock.Anything, notifID).
		Return(errors.New("deletion failed"))

	err := svc.RejectNotification(context.Background(), notifID)

	assert.Error(t, err)
	assert.Equal(t, "deletion failed", err.Error())
}

func TestAcceptPartyInvite_Success(t *testing.T) {
	svc := setupSocialService()

	partyInviteReq := models.PartyInviteRequest{
		PartyID:     uuid.NewString(),
		SenderID:    "user123",
		RecipientID: "user456",
		MemberCount: "2",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("AddPartyMember", mock.Anything, partyInviteReq).
		Return(nil)

	err := svc.AcceptPartyInvite(context.Background(), partyInviteReq)

	assert.NoError(t, err)
	svc.socialrepo.(*socialMocks.MockSocialRepository).AssertExpectations(t)
}

func TestAcceptPartyInvite_Error(t *testing.T) {
	svc := setupSocialService()

	partyInviteReq := models.PartyInviteRequest{
		PartyID:     uuid.NewString(),
		SenderID:    "user123",
		RecipientID: "user456",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("AddPartyMember", mock.Anything, partyInviteReq).
		Return(errors.New("party invite not found"))

	err := svc.AcceptPartyInvite(context.Background(), partyInviteReq)

	assert.Error(t, err)
	assert.Equal(t, "party invite not found", err.Error())
}

func TestAcceptNotification_FriendRequest_Success(t *testing.T) {
	svc := setupSocialService()

	notif := models.Notification{
		ID:          uuid.NewString(),
		SenderID:    "user123",
		RecipientID: "user456",
		Type:        models.FriendRequest,
		Status:      "Pending",
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("AddFriend", mock.Anything, notif.SenderID, notif.RecipientID).
		Return(nil)

	svc.notificationservice.(*notificationMocks.MockNotificationService).
		On("CreateAndPublishNotification", mock.Anything, mock.MatchedBy(func(n models.Notification) bool {
			return n.Type == models.FriendRequest
		})).
		Return(nil)

	err := svc.AcceptNotification(context.Background(), notif)

	assert.NoError(t, err)
	svc.socialrepo.(*socialMocks.MockSocialRepository).AssertExpectations(t)
}

func TestAcceptNotification_FriendRequest_AddFriendError(t *testing.T) {
	svc := setupSocialService()

	notif := models.Notification{
		ID:          uuid.NewString(),
		SenderID:    "user123",
		RecipientID: "user456",
		Type:        models.FriendRequest,
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("AddFriend", mock.Anything, notif.SenderID, notif.RecipientID).
		Return(errors.New("friend addition failed"))

	err := svc.AcceptNotification(context.Background(), notif)

	assert.Error(t, err)
	assert.Equal(t, "friend addition failed", err.Error())
}

func TestAcceptNotification_FriendRequest_PublishError(t *testing.T) {
	svc := setupSocialService()

	notif := models.Notification{
		ID:          uuid.NewString(),
		SenderID:    "user123",
		RecipientID: "user456",
		Type:        models.FriendRequest,
	}

	svc.socialrepo.(*socialMocks.MockSocialRepository).
		On("AddFriend", mock.Anything, notif.SenderID, notif.RecipientID).
		Return(nil)

	svc.notificationservice.(*notificationMocks.MockNotificationService).
		On("CreateAndPublishNotification", mock.Anything, mock.Anything).
		Return(errors.New("publish failed"))

	err := svc.AcceptNotification(context.Background(), notif)

	assert.Error(t, err)
	assert.Equal(t, "publish failed", err.Error())
}

func TestAcceptNotification_UnsupportedType(t *testing.T) {
	svc := setupSocialService()

	notif := models.Notification{
		ID:          uuid.NewString(),
		SenderID:    "user123",
		RecipientID: "user456",
		Type:        models.MatchFound, // Unsupported type for social service
	}

	err := svc.AcceptNotification(context.Background(), notif)

	// Should not error, just return nil for unhandled types
	assert.NoError(t, err)
}
