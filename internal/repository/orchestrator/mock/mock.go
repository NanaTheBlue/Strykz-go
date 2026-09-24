package mock

import (
	"context"
	"time"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockOrchestratorRepo struct {
	mock.Mock
}

func (m *MockOrchestratorRepo) UpdateHeartBeat(ctx context.Context, serverid string) error {
	args := m.Called(ctx, serverid)
	return args.Error(0)
}

func (m *MockOrchestratorRepo) GetDeadServers(ctx context.Context, cutoff time.Time) ([]models.Gameserver, error) {
	args := m.Called(ctx, cutoff)
	if servers := args.Get(0); servers != nil {
		return servers.([]models.Gameserver), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrchestratorRepo) GetServersByRegion(ctx context.Context, region string) ([]models.Gameserver, error) {
	args := m.Called(ctx, region)
	if servers := args.Get(0); servers != nil {
		return servers.([]models.Gameserver), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrchestratorRepo) InsertServer(ctx context.Context, server models.Gameserver) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockOrchestratorRepo) DeleteServer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrchestratorRepo) UpdateServer(ctx context.Context, id string, status models.ServerStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrchestratorRepo) AcquireReadyServer(ctx context.Context, region string) (*models.Gameserver, error) {
	args := m.Called(ctx, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Gameserver), args.Error(1)
}

func (m *MockOrchestratorRepo) CountReadyServers(ctx context.Context, region string) (int, error) {
	args := m.Called(ctx, region)
	return args.Int(0), args.Error(1)
}
