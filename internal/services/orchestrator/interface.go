package orchestrator

import "context"

type Service interface {
	UpdateHeartbeat(ctx context.Context, serverID string) error
	CreateServer(ctx context.Context, region string) (string, error)
}
