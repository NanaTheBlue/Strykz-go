package orchestrator

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Service interface {
	UpdateHeartbeat(ctx context.Context, serverID string) error
	CreateServer(ctx context.Context, region string) (string, error)
	Request(region string)
	UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error
}
