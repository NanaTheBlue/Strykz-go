package orchestrator

import "context"

type Service interface {
	UpdateHeartbeat(serverID string, ctx context.Context) error
	CreateServer(ctx context.Context, region string) (string, error)
}
