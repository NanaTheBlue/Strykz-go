//go:build unit

package notifications

import (
	"context"
	"errors"
	"testing"

	notifmock "github.com/nanagoboiler/internal/repository/notification/mock"
	redismock "github.com/nanagoboiler/internal/repository/redis/mock"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAndPublishNotification_Success(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	mockRedis := new(redismock.MockRedisStore)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		store:            mockRedis,
		notificationrepo: mockNotifRepo,
	}

	notif := models.Notification{
		ID:          "notif-1",
		SenderID:    "sender",
		RecipientID: "recipient",
		Type:        models.FriendRequest,
	}

	mockNotifRepo.On("SendNotification", mock.Anything, notif).Return("notif-1", nil)
	mockRedis.On("Publish", mock.Anything, "notifications", notif).Return(nil)

	err := svc.CreateAndPublishNotification(context.Background(), notif)
	assert.NoError(t, err)
	mockNotifRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestCreateAndPublishNotification_RepoError(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	mockRedis := new(redismock.MockRedisStore)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		store:            mockRedis,
		notificationrepo: mockNotifRepo,
	}

	notif := models.Notification{ID: "notif-1"}
	mockNotifRepo.On("SendNotification", mock.Anything, notif).Return("", errors.New("db error"))

	err := svc.CreateAndPublishNotification(context.Background(), notif)
	assert.Error(t, err)
}

func TestCreateNoPublishNotification_Success(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		notificationrepo: mockNotifRepo,
	}

	notif := models.Notification{ID: "notif-1"}
	mockNotifRepo.On("SendNotification", mock.Anything, notif).Return("notif-1", nil)

	id, err := svc.CreateNoPublishNotification(context.Background(), notif)
	assert.NoError(t, err)
	assert.Equal(t, "notif-1", id)
}

func TestCreateNoPublishNotification_RepoError(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		notificationrepo: mockNotifRepo,
	}

	notif := models.Notification{ID: "notif-1"}
	mockNotifRepo.On("SendNotification", mock.Anything, notif).Return("", errors.New("db error"))

	_, err := svc.CreateNoPublishNotification(context.Background(), notif)
	assert.Error(t, err)
}

func TestGetNotifications_Success(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		notificationrepo: mockNotifRepo,
	}

	expected := []models.Notification{
		{ID: "n1", RecipientID: "user1"},
		{ID: "n2", RecipientID: "user1"},
	}
	mockNotifRepo.On("GetNotifications", mock.Anything, "user1").Return(expected, nil)

	notifs, err := svc.GetNotifications(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Len(t, notifs, 2)
}

func TestGetNotifications_Empty(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		notificationrepo: mockNotifRepo,
	}

	mockNotifRepo.On("GetNotifications", mock.Anything, "user1").Return([]models.Notification{}, nil)

	notifs, err := svc.GetNotifications(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Len(t, notifs, 0)
}

func TestDeleteNotification_Success(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		notificationrepo: mockNotifRepo,
	}

	mockNotifRepo.On("DeleteNotification", mock.Anything, "notif-1").Return(nil)

	err := svc.DeleteNotification(context.Background(), "notif-1")
	assert.NoError(t, err)
}

func TestDeleteNotification_Error(t *testing.T) {
	mockNotifRepo := new(notifmock.MockNotificationRepo)
	hub := NewHub()

	svc := &notificationsService{
		hub:              hub,
		notificationrepo: mockNotifRepo,
	}

	mockNotifRepo.On("DeleteNotification", mock.Anything, "notif-1").Return(errors.New("not found"))

	err := svc.DeleteNotification(context.Background(), "notif-1")
	assert.Error(t, err)
}

func TestHub_Send_NoConnection(t *testing.T) {
	hub := NewHub()
	err := hub.Send("nonexistent", models.Notification{})
	assert.NoError(t, err)
}

func TestNewHub(t *testing.T) {
	hub := NewHub()
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.connections)
}

func TestNotificationsService_AddConnection(t *testing.T) {
	hub := NewHub()
	svc := &notificationsService{hub: hub}
	svc.AddConnection("user1", nil)
}

func TestNotificationsService_RemoveConnection(t *testing.T) {
	hub := NewHub()
	svc := &notificationsService{hub: hub}
	svc.RemoveConnection("user1")
}

func TestPublishNotification_Success(t *testing.T) {
	mockRedis := new(redismock.MockRedisStore)
	hub := NewHub()

	svc := &notificationsService{
		hub:   hub,
		store: mockRedis,
	}

	notif := models.Notification{ID: "n1"}
	mockRedis.On("Publish", mock.Anything, "notifications", notif).Return(nil)

	err := svc.PublishNotification(context.Background(), notif)
	assert.NoError(t, err)
}

func TestPublishNotification_Error(t *testing.T) {
	mockRedis := new(redismock.MockRedisStore)
	hub := NewHub()

	svc := &notificationsService{
		hub:   hub,
		store: mockRedis,
	}

	notif := models.Notification{ID: "n1"}
	mockRedis.On("Publish", mock.Anything, "notifications", notif).Return(errors.New("redis error"))

	err := svc.PublishNotification(context.Background(), notif)
	assert.Error(t, err)
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	hub.Broadcast(models.Notification{})
}

func TestHub_Remove_NoConnection(t *testing.T) {
	hub := NewHub()
	hub.Remove("nonexistent")
}
