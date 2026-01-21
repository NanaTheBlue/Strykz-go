package orchestrator

import (
	"context"

	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
)

type Orchestrator struct {
	orchestratorrepo orchestratorrepo.OrchestratoryRepository
}

func NewOrchestrator(orchestratorrepo orchestratorrepo.OrchestratoryRepository) Service {
	return &Orchestrator{
		orchestratorrepo: orchestratorrepo,
	}
}

func (s *Orchestrator) UpdateHeartbeat(serverID string, ctx context.Context) error {

	//TODO: Better Error Handling
	return s.orchestratorrepo.UpdateHeartBeat(serverID, ctx)
}
