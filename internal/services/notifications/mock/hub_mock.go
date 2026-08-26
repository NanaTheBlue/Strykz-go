package mock

import (
	"github.com/gorilla/websocket"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockHub struct {
	mock.Mock
}

func (m *MockHub) Add(userID string, conn *websocket.Conn) {
	m.Called(userID, conn)
}

func (m *MockHub) Remove(userID string) {
	m.Called(userID)
}

func (m *MockHub) Send(userID string, notif models.Notification) error {
	args := m.Called(userID, notif)
	return args.Error(0)
}
