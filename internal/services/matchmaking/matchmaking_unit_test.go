//go:build unit

package matchmaking

import (
	"context"
	"errors"
	"testing"
	"time"

	redismock "github.com/nanagoboiler/internal/repository/redis/mock"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) CreateNoPublishNotification(ctx context.Context, notif models.Notification) (string, error) {
	args := m.Called(ctx, notif)
	return args.String(0), args.Error(1)
}

type MockCapacityRequester struct {
	mock.Mock
}

func (m *MockCapacityRequester) Request(region string) {
	m.Called(region)
}

type MockServerSpeaker struct {
	mock.Mock
}

func (m *MockServerSpeaker) ReloadWhitelist(serverID string, steamIDs []string) error {
	args := m.Called(serverID, steamIDs)
	return args.Error(0)
}

type MockMatchmakingRepo struct {
	mock.Mock
}

func (m *MockMatchmakingRepo) InsertPlayers(ctx context.Context, players []*models.Player, matchid string) error {
	args := m.Called(ctx, players, matchid)
	return args.Error(0)
}

func (m *MockMatchmakingRepo) CreateMatch(ctx context.Context, deadline time.Time, region string) (string, error) {
	args := m.Called(ctx, deadline, region)
	return args.String(0), args.Error(1)
}

func (m *MockMatchmakingRepo) UpdatePlayer(ctx context.Context, player models.Player, matchid string, status string) error {
	args := m.Called(ctx, player, matchid, status)
	return args.Error(0)
}

func (m *MockMatchmakingRepo) AreAllPlayersAccepted(ctx context.Context, matchID string) (bool, error) {
	args := m.Called(ctx, matchID)
	return args.Bool(0), args.Error(1)
}

func (m *MockMatchmakingRepo) UpdateMatchStatus(ctx context.Context, matchID string, status models.MatchStatus) error {
	args := m.Called(ctx, matchID, status)
	return args.Error(0)
}

func (m *MockMatchmakingRepo) AssignServerToMatch(ctx context.Context, matchID string, serverID string) error {
	args := m.Called(ctx, matchID, serverID)
	return args.Error(0)
}

func (m *MockMatchmakingRepo) GetMatch(ctx context.Context, matchID string) (models.Match, error) {
	args := m.Called(ctx, matchID)
	return args.Get(0).(models.Match), args.Error(1)
}

func (m *MockMatchmakingRepo) GetMatchPlayers(ctx context.Context, matchID string) ([]models.Player, error) {
	args := m.Called(ctx, matchID)
	if p := args.Get(0); p != nil {
		return p.([]models.Player), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMatchmakingRepo) GetMatchesByStatus(ctx context.Context, status models.MatchStatus) ([]models.Match, error) {
	args := m.Called(ctx, status)
	if m := args.Get(0); m != nil {
		return m.([]models.Match), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMatchmakingRepo) GetPlayerByID(ctx context.Context, userID string) (models.Player, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(models.Player), args.Error(1)
}

func TestInQue_Success(t *testing.T) {
	mockRedis := new(redismock.MockRedisStore)
	svc := &matchmakingService{
		RedisRepo: mockRedis,
	}

	player := &models.Player{
		Player_id:      "player-1",
		Player_steamid: "steam-1",
	}

	mockRedis.On("Que", mock.Anything, "1v1", "us", player).Return(nil)

	err := svc.InQue(context.Background(), player)
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestInQue_RedisError(t *testing.T) {
	mockRedis := new(redismock.MockRedisStore)
	svc := &matchmakingService{
		RedisRepo: mockRedis,
	}

	player := &models.Player{Player_id: "player-1"}
	mockRedis.On("Que", mock.Anything, "1v1", "us", player).Return(errors.New("redis down"))

	err := svc.InQue(context.Background(), player)
	assert.Error(t, err)
}

func TestGetPlayerByID_Success(t *testing.T) {
	mockRepo := new(MockMatchmakingRepo)
	svc := &matchmakingService{
		matchmakingrepo: mockRepo,
	}

	expected := models.Player{
		Player_id:      "user-1",
		Player_steamid: "steam-1",
	}
	mockRepo.On("GetPlayerByID", mock.Anything, "user-1").Return(expected, nil)

	player, err := svc.GetPlayerByID(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Equal(t, expected, player)
}

func TestGetPlayerByID_NotFound(t *testing.T) {
	mockRepo := new(MockMatchmakingRepo)
	svc := &matchmakingService{
		matchmakingrepo: mockRepo,
	}

	mockRepo.On("GetPlayerByID", mock.Anything, "missing").Return(models.Player{}, errors.New("not found"))

	player, err := svc.GetPlayerByID(context.Background(), "missing")
	assert.Error(t, err)
	assert.Equal(t, models.Player{}, player)
}

func TestQueReader_InsufficientPlayers(t *testing.T) {
	mockRedis := new(redismock.MockRedisStore)
	svc := &matchmakingService{
		RedisRepo: mockRedis,
	}

	mockRedis.On("DeQue", mock.Anything, "1v1", "us", 2).Return([]*models.Player{}, nil)

	svc.QueReader(context.Background(), "1v1")
	mockRedis.AssertExpectations(t)
}

func TestQueReader_EnoughPlayers(t *testing.T) {
	mockRedis := new(redismock.MockRedisStore)
	svc := &matchmakingService{
		RedisRepo: mockRedis,
	}

	players := []*models.Player{
		{Player_id: "p1", Player_steamid: "s1"},
		{Player_id: "p2", Player_steamid: "s2"},
	}
	mockRedis.On("DeQue", mock.Anything, "1v1", "us", 2).Return(players, nil)

	svc.QueReader(context.Background(), "1v1")
}

var _ Notifier = (*MockNotifier)(nil)
var _ CapacityRequester = (*MockCapacityRequester)(nil)
var _ ServerSpeaker = (*MockServerSpeaker)(nil)
var _ Service = (*matchmakingService)(nil)
