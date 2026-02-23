package grpcserver_test

import (
	"context"

	"net"
	"testing"
	"time"

	pb "github.com/nanagoboiler/gen"
	"github.com/stretchr/testify/require"

	grpcserver "github.com/nanagoboiler/internal/grpc"
	"github.com/nanagoboiler/internal/services/orchestrator"
	"github.com/nanagoboiler/models"
	"google.golang.org/grpc"
)

type fakeRepo struct{}

func (f *fakeRepo) UpdateHeartBeat(ctx context.Context, id string) error { return nil }
func (f *fakeRepo) AcquireReadyServer(ctx context.Context, region string) (*models.Gameserver, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateServer(ctx context.Context, id string, status models.ServerStatus) error {
	return nil
}
func (f *fakeRepo) InsertServer(ctx context.Context, server models.Gameserver) error {
	return nil
}
func (f *fakeRepo) DeleteServer(ctx context.Context, serverid string) error {
	return nil
}
func (f *fakeRepo) CountReadyServers(ctx context.Context, region string) (int, error) {
	return 0, nil
}

func (f *fakeRepo) GetDeadServers(ctx context.Context, deadline time.Time) ([]models.Gameserver, error) {
	return []models.Gameserver{}, nil
}
func (f *fakeRepo) GetServersByRegion(ctx context.Context, region string) ([]models.Gameserver, error) {
	return []models.Gameserver{}, nil
}

//Todo: clean ^ this up a bit

func startGrpcServer(t *testing.T) (*grpc.Server, net.Listener, *orchestrator.Orchestrator) {
	repo := &fakeRepo{}
	orch := orchestrator.NewOrchestrator(repo, nil).(*orchestrator.Orchestrator)

	lis, err := net.Listen("tcp", ":6767")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer()

	pb.RegisterSidecarServiceServer(
		server,
		grpcserver.NewSidecarServer(orch),
	)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("grpc server stopped: %v", err)
		}
	}()

	return server, lis, orch
}

func TestReloadWhitelistIntegration(t *testing.T) {
	server, lis, orch := startGrpcServer(t)
	defer server.Stop()
	defer lis.Close()

	serverID := "68b434d1-e394-41bb-9e00-85b50432f1b0"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deadline := time.Now().Add(10 * time.Second)
	for orch.GetStream(serverID) == nil {
		if time.Now().After(deadline) || ctx.Err() != nil {
			t.Fatal("sidecar did not connect")
		}
		time.Sleep(50 * time.Millisecond)
	}

	err := orch.ReloadWhitelist(serverID, []string{"123"})

	require.NoError(t, err)

	stream := orch.GetStream(serverID)
	require.NotNil(t, stream, "stream disappeared unexpectedly")

}
