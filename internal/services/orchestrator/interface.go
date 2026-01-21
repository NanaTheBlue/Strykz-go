package orchestrator

import "context"

type Service interface {
	UpdateHeartbeat(serverID string, ctx context.Context) error
}
