package mock

import (
	"context"
	"time"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

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
	players, _ := args.Get(0).([]models.Player)
	return players, args.Error(1)
}

func (m *MockMatchmakingRepo) GetMatchesByStatus(ctx context.Context, status models.MatchStatus) ([]models.Match, error) {
	args := m.Called(ctx, status)
	matches, _ := args.Get(0).([]models.Match)
	return matches, args.Error(1)
}

func (m *MockMatchmakingRepo) GetPlayerByID(ctx context.Context, userID string) (models.Player, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(models.Player), args.Error(1)
}
