package mock

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) StartBackgroundListener(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockNotificationService) AddConnection(userID string, conn *websocket.Conn) {
	m.Called(userID, conn)
}

func (m *MockNotificationService) RemoveConnection(userID string) {
	m.Called(userID)
}

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
	var notifications []models.Notification
	if n := args.Get(0); n != nil {
		notifications = n.([]models.Notification)
	}
	return notifications, args.Error(1)
}

func (m *MockNotificationService) DeleteNotification(ctx context.Context, notifID string) error {
	args := m.Called(ctx, notifID)
	return args.Error(0)
}
