package orchestrator

import (
	"context"

	pb "github.com/nanagoboiler/gen"
	"github.com/nanagoboiler/models"
)

type Service interface {
	UpdateHeartbeat(ctx context.Context, serverID string) error
	CreateServer(ctx context.Context, region string) (string, error)
	Request(region string)
	UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error
	RegisterStream(serverID string, stream pb.SidecarService_ConnectServer)
	UnregisterStream(serverID string)
	GetStream(serverID string) pb.SidecarService_ConnectServer
	SelectServer(ctx context.Context, region string) (*models.Gameserver, error)
	ReloadWhitelist(serverID string, steamIDs []string) error
}
