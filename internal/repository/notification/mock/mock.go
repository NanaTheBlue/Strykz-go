package mock

import (
	"context"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockNotificationRepo struct {
	mock.Mock
}

func (m *MockNotificationRepo) GetNotifications(ctx context.Context, uuid string) ([]models.Notification, error) {
	args := m.Called(ctx, uuid)
	notifs, _ := args.Get(0).([]models.Notification)
	return notifs, args.Error(1)
}

func (m *MockNotificationRepo) GetNotification(ctx context.Context, notificationID string) (models.Notification, error) {
	args := m.Called(ctx, notificationID)
	return args.Get(0).(models.Notification), args.Error(1)
}

func (m *MockNotificationRepo) SendNotification(ctx context.Context, notif models.Notification) (string, error) {
	args := m.Called(ctx, notif)
	return args.String(0), args.Error(1)
}

func (m *MockNotificationRepo) DeleteNotification(ctx context.Context, notificationID string) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}
