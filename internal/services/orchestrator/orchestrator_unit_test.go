//go:build unit

package orchestrator

import (
	"context"
	"testing"

	pb "github.com/nanagoboiler/gen"
	orchmock "github.com/nanagoboiler/internal/repository/orchestrator/mock"
	gameserverconfig "github.com/nanagoboiler/internal/services/config"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestOrchestrator() (*Orchestrator, *orchmock.MockOrchestratorRepo) {
	mockRepo := new(orchmock.MockOrchestratorRepo)
	cfg := gameserverconfig.Config{
		AMI:           "ami-test",
		SubnetID:      "subnet-test",
		SecurityGroup: "sg-test",
		InstanceType:  "t3.micro",
	}
	o := &Orchestrator{
		orchestratorrepo: mockRepo,
		streams:          make(map[string]pb.SidecarService_ConnectServer),
		cfg:              cfg,
	}
	return o, mockRepo
}

func TestUpdateHeartbeat_Success(t *testing.T) {
	o, mockRepo := newTestOrchestrator()
	ctx := context.Background()

	mockRepo.On("UpdateHeartBeat", ctx, "server-1").Return(nil)

	err := o.UpdateHeartbeat(ctx, "server-1")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateHeartbeat_Error(t *testing.T) {
	o, mockRepo := newTestOrchestrator()
	ctx := context.Background()

	mockRepo.On("UpdateHeartBeat", ctx, "server-1").Return(assert.AnError)

	err := o.UpdateHeartbeat(ctx, "server-1")
	assert.Error(t, err)
}

func TestUpdateServerStatus_Success(t *testing.T) {
	o, mockRepo := newTestOrchestrator()
	ctx := context.Background()

	mockRepo.On("UpdateServer", ctx, "server-1", models.ServerReady).Return(nil)

	err := o.UpdateServerStatus(ctx, "server-1", models.ServerReady)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateServerStatus_Error(t *testing.T) {
	o, mockRepo := newTestOrchestrator()
	ctx := context.Background()

	mockRepo.On("UpdateServer", ctx, "server-1", models.ServerBusy).Return(assert.AnError)

	err := o.UpdateServerStatus(ctx, "server-1", models.ServerBusy)
	assert.Error(t, err)
}

func TestRegisterAndUnregisterStream(t *testing.T) {
	o, _ := newTestOrchestrator()

	o.RegisterStream("server-1", nil)
	stream := o.GetStream("server-1")
	assert.Nil(t, stream)

	o.UnregisterStream("server-1")
	stream = o.GetStream("server-1")
	assert.Nil(t, stream)
}

func TestGetStream_NotFound(t *testing.T) {
	o, _ := newTestOrchestrator()

	stream := o.GetStream("nonexistent")
	assert.Nil(t, stream)
}

func TestRequest_NoReadyServers(t *testing.T) {
	o, mockRepo := newTestOrchestrator()

	mockRepo.On("CountReadyServers", mock.Anything, "us").Return(1, nil)

	o.Request("us")

	// Give goroutine time to execute
}

func TestRequest_CountError(t *testing.T) {
	o, mockRepo := newTestOrchestrator()

	mockRepo.On("CountReadyServers", mock.Anything, "us").Return(0, assert.AnError)

	o.Request("us")
}

func TestRequest_HasReadyServers(t *testing.T) {
	o, mockRepo := newTestOrchestrator()

	mockRepo.On("CountReadyServers", mock.Anything, "us").Return(2, nil)

	o.Request("us")
}

func TestSelectServer_Found(t *testing.T) {
	o, mockRepo := newTestOrchestrator()
	ctx := context.Background()

	expected := &models.Gameserver{ID: "server-1", Region: "us", Status: models.ServerReady}
	mockRepo.On("AcquireReadyServer", ctx, "us").Return(expected, nil)

	server, err := o.SelectServer(ctx, "us")
	assert.NoError(t, err)
	assert.Equal(t, expected, server)
}

func TestSelectServer_NotFound(t *testing.T) {
	o, mockRepo := newTestOrchestrator()
	ctx := context.Background()

	var nilServer *models.Gameserver
	mockRepo.On("AcquireReadyServer", ctx, "us").Return(nilServer, nil)

	server, err := o.SelectServer(ctx, "us")
	assert.NoError(t, err)
	assert.Nil(t, server)
}

func TestSelectServer_Error(t *testing.T) {
	o, mockRepo := newTestOrchestrator()
	ctx := context.Background()

	mockRepo.On("AcquireReadyServer", ctx, "us").Return((*models.Gameserver)(nil), assert.AnError)

	server, err := o.SelectServer(ctx, "us")
	assert.Error(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, "", server.ID)
}

func TestReloadWhitelist_NoStream(t *testing.T) {
	o, _ := newTestOrchestrator()

	err := o.ReloadWhitelist("nonexistent-server", []string{"steam1", "steam2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no stream")
}
