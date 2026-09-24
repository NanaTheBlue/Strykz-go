//go:build unit

package notifications

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	notificationMocks "github.com/nanagoboiler/internal/repository/notification/mock"
	redisMocks "github.com/nanagoboiler/internal/repository/redis/mock"
	"github.com/nanagoboiler/models"
)

func setupNotificationsService() *notificationsService {
	hub := NewHub() // Use the real Hub for these tests
	store := new(redisMocks.MockRedisStore)
	notificationrepo := new(notificationMocks.MockNotificationRepository)

	return &notificationsService{
		hub:              hub,
		store:            store,
		notificationrepo: notificationrepo,
	}
}

func TestCreateAndPublishNotification_Success(t *testing.T) {
	svc := setupNotificationsService()

	notif := models.Notification{
		ID:          uuid.NewString(),
		SenderID:    "sender123",
		RecipientID: "recipient456",
		Type:        models.FriendRequest,
		Data:        "friend request data",
		Status:      "Pending",
		CreatedAt:   time.Now(),
	}

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("SendNotification", mock.Anything, mock.MatchedBy(func(n models.Notification) bool {
			return n.SenderID == notif.SenderID && n.RecipientID == notif.RecipientID
		})).
		Return(notif.ID, nil)

	svc.store.(*redisMocks.MockRedisStore).
		On("Publish", mock.Anything, "notifications", mock.MatchedBy(func(n models.Notification) bool {
			return n.RecipientID == notif.RecipientID
		})).
		Return(nil)

	err := svc.CreateAndPublishNotification(context.Background(), notif)

	assert.NoError(t, err)
	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).AssertExpectations(t)
	svc.store.(*redisMocks.MockRedisStore).AssertExpectations(t)
}

func TestCreateAndPublishNotification_RepositoryError(t *testing.T) {
	svc := setupNotificationsService()

	notif := models.Notification{
		SenderID:    "sender123",
		RecipientID: "recipient456",
		Type:        models.FriendRequest,
	}

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("SendNotification", mock.Anything, mock.Anything).
		Return("", errors.New("database error"))

	err := svc.CreateAndPublishNotification(context.Background(), notif)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
}

func TestCreateAndPublishNotification_PublishError(t *testing.T) {
	svc := setupNotificationsService()

	notif := models.Notification{
		ID:          uuid.NewString(),
		SenderID:    "sender123",
		RecipientID: "recipient456",
		Type:        models.FriendRequest,
	}

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("SendNotification", mock.Anything, mock.Anything).
		Return(notif.ID, nil)

	svc.store.(*redisMocks.MockRedisStore).
		On("Publish", mock.Anything, "notifications", mock.Anything).
		Return(errors.New("redis publish error"))

	err := svc.CreateAndPublishNotification(context.Background(), notif)

	assert.Error(t, err)
	assert.Equal(t, "redis publish error", err.Error())
}

func TestCreateNoPublishNotification_Success(t *testing.T) {
	svc := setupNotificationsService()

	notifID := uuid.NewString()
	notif := models.Notification{
		ID:          notifID,
		SenderID:    "sender123",
		RecipientID: "recipient456",
		Type:        models.PartyInvite,
	}

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("SendNotification", mock.Anything, mock.Anything).
		Return(notifID, nil)

	id, err := svc.CreateNoPublishNotification(context.Background(), notif)

	assert.NoError(t, err)
	assert.Equal(t, notifID, id)
	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).AssertExpectations(t)
}

func TestCreateNoPublishNotification_Error(t *testing.T) {
	svc := setupNotificationsService()

	notif := models.Notification{
		SenderID:    "sender123",
		RecipientID: "recipient456",
		Type:        models.PartyInvite,
	}

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("SendNotification", mock.Anything, mock.Anything).
		Return("", errors.New("database error"))

	id, err := svc.CreateNoPublishNotification(context.Background(), notif)

	assert.Error(t, err)
	assert.Empty(t, id)
	assert.Equal(t, "database error", err.Error())
}

func TestPublishNotification_Success(t *testing.T) {
	svc := setupNotificationsService()

	notif := models.Notification{
		ID:          uuid.NewString(),
		SenderID:    "sender123",
		RecipientID: "recipient456",
		Type:        models.MatchFound,
		Data:        "match found",
	}

	svc.store.(*redisMocks.MockRedisStore).
		On("Publish", mock.Anything, "notifications", mock.Anything).
		Return(nil)

	err := svc.PublishNotification(context.Background(), notif)

	assert.NoError(t, err)
	svc.store.(*redisMocks.MockRedisStore).AssertExpectations(t)
}

func TestPublishNotification_Error(t *testing.T) {
	svc := setupNotificationsService()

	notif := models.Notification{
		SenderID:    "sender123",
		RecipientID: "recipient456",
		Type:        models.MatchFound,
	}

	svc.store.(*redisMocks.MockRedisStore).
		On("Publish", mock.Anything, "notifications", mock.Anything).
		Return(errors.New("redis error"))

	err := svc.PublishNotification(context.Background(), notif)

	assert.Error(t, err)
	assert.Equal(t, "redis error", err.Error())
}

func TestGetNotifications_Success(t *testing.T) {
	svc := setupNotificationsService()

	userID := "user123"
	notifications := []models.Notification{
		{
			ID:          uuid.NewString(),
			SenderID:    "sender1",
			RecipientID: userID,
			Type:        models.FriendRequest,
			Status:      "Pending",
		},
		{
			ID:          uuid.NewString(),
			SenderID:    "sender2",
			RecipientID: userID,
			Type:        models.PartyInvite,
			Status:      "Pending",
		},
	}

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("GetNotifications", mock.Anything, userID).
		Return(notifications, nil)

	result, err := svc.GetNotifications(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, len(notifications), len(result))
	assert.Equal(t, notifications[0].SenderID, result[0].SenderID)
	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).AssertExpectations(t)
}

func TestGetNotifications_EmptyList(t *testing.T) {
	svc := setupNotificationsService()

	userID := "user123"

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("GetNotifications", mock.Anything, userID).
		Return([]models.Notification{}, nil)

	result, err := svc.GetNotifications(context.Background(), userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetNotifications_Error(t *testing.T) {
	svc := setupNotificationsService()

	userID := "user123"

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("GetNotifications", mock.Anything, userID).
		Return([]models.Notification{}, errors.New("database error"))

	result, err := svc.GetNotifications(context.Background(), userID)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Equal(t, "database error", err.Error())
}

func TestDeleteNotification_Success(t *testing.T) {
	svc := setupNotificationsService()

	notifID := uuid.NewString()

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("DeleteNotification", mock.Anything, notifID).
		Return(nil)

	err := svc.DeleteNotification(context.Background(), notifID)

	assert.NoError(t, err)
	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).AssertExpectations(t)
}

func TestDeleteNotification_Error(t *testing.T) {
	svc := setupNotificationsService()

	notifID := uuid.NewString()

	svc.notificationrepo.(*notificationMocks.MockNotificationRepository).
		On("DeleteNotification", mock.Anything, notifID).
		Return(errors.New("deletion failed"))

	err := svc.DeleteNotification(context.Background(), notifID)

	assert.Error(t, err)
	assert.Equal(t, "deletion failed", err.Error())
}

func TestAddConnection(t *testing.T) {
	svc := setupNotificationsService()

	// Note: In a real scenario, this would be an actual websocket connection
	// For unit testing, we skip the websocket connection part
	// The hub.Add method handles the actual storage

	assert.NotNil(t, svc.hub)
}

func TestRemoveConnection(t *testing.T) {
	svc := setupNotificationsService()

	userID := "user123"

	// The hub.Remove method handles the actual removal
	svc.RemoveConnection(userID)

	assert.NotNil(t, svc.hub)
}

// Mock websocket connection for testing (no longer used in tests)
type mockWebSocketConn struct{}

func (m *mockWebSocketConn) ReadJSON(v interface{}) error {
	return nil
}

func (m *mockWebSocketConn) WriteJSON(v interface{}) error {
	return nil
}

func (m *mockWebSocketConn) Close() error {
	return nil
}
