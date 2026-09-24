package mock

import (
	"context"
	"time"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockRedisStore struct {
	mock.Mock
}

func (m *MockRedisStore) Add(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisStore) AddNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockRedisStore) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisStore) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedisStore) Subscribe(ctx context.Context, channel string, handler func(message string)) error {
	args := m.Called(ctx, channel, handler)
	return args.Error(0)
}

func (m *MockRedisStore) Publish(ctx context.Context, channel string, message models.Notification) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

func (m *MockRedisStore) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockRedisStore) Count(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisStore) Que(ctx context.Context, mode, region string, player *models.Player) error {
	args := m.Called(ctx, mode, region, player)
	return args.Error(0)
}

func (m *MockRedisStore) DeQue(ctx context.Context, mode, region string, count int) ([]*models.Player, error) {
	args := m.Called(ctx, mode, region, count)
	var players []*models.Player
	if p := args.Get(0); p != nil {
		players = p.([]*models.Player)
	}
	return players, args.Error(1)
}

func (m *MockRedisStore) DeQuePlayer(ctx context.Context, mode, region, playerID string) error {
	args := m.Called(ctx, mode, region, playerID)
	return args.Error(0)
}
