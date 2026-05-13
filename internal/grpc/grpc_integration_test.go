//go:build integration

package grpcserver_test

import (
	"context"

	"net"
	"testing"
	"time"

	pb "github.com/nanagoboiler/gen"

	grpcserver "github.com/nanagoboiler/internal/grpc"
	gameserverconfig "github.com/nanagoboiler/internal/services/config"
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

func newTestOrchestrator() orchestrator.Service {
	cfg := gameserverconfig.Config{
		AMI:           "ami-test",
		SubnetID:      "subnet-test",
		SecurityGroup: "sg-test",
		InstanceType:  "t3.micro",
	}

	return orchestrator.NewOrchestrator(&fakeRepo{}, nil, cfg)
}
func startGrpcServer(t *testing.T) (*grpc.Server, net.Listener, orchestrator.Service) {
	orch := newTestOrchestrator()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer()

	pb.RegisterSidecarServiceServer(
		server,
		grpcserver.NewSidecarServer(orch),
	)

	go server.Serve(lis)

	t.Cleanup(func() {
		server.Stop()
		lis.Close()
	})

	return server, lis, orch
}
