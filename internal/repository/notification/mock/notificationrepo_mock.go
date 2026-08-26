package mock

import (
	"context"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) SendNotification(ctx context.Context, notif models.Notification) (string, error) {
	args := m.Called(ctx, notif)
	return args.String(0), args.Error(1)
}

func (m *MockNotificationRepository) DeleteNotification(ctx context.Context, notifID string) error {
	args := m.Called(ctx, notifID)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotifications(ctx context.Context, userID string) ([]models.Notification, error) {
	args := m.Called(ctx, userID)
	var notifications []models.Notification
	if n := args.Get(0); n != nil {
		notifications = n.([]models.Notification)
	}
	return notifications, args.Error(1)
}

func (m *MockNotificationRepository) GetNotification(ctx context.Context, notificationID string) (models.Notification, error) {
	args := m.Called(ctx, notificationID)
	var notif models.Notification
	if n := args.Get(0); n != nil {
		notif = n.(models.Notification)
	}
	return notif, args.Error(1)
}
