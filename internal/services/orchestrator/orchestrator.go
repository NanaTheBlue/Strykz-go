package orchestrator

import (
	"context"

	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	"github.com/nanagoboiler/models"
	"github.com/vultr/govultr/v3"
)

type Orchestrator struct {
	orchestratorrepo orchestratorrepo.OrchestratoryRepository
	vultrclient      govultr.Client
}

func NewOrchestrator(orchestratorrepo orchestratorrepo.OrchestratoryRepository, vultrclient govultr.Client) Service {
	return &Orchestrator{
		orchestratorrepo: orchestratorrepo,
		vultrclient:      vultrclient,
	}
}

func (s *Orchestrator) UpdateHeartbeat(serverID string, ctx context.Context) error {

	//TODO: Better Error Handling
	return s.orchestratorrepo.UpdateHeartBeat(serverID, ctx)
}

func (s *Orchestrator) SelectServer(ctx context.Context, region string) (models.Gameserver, error) {
	//TODO: Better Error Handling honestly i should make this a github issue

	Gameserver, err := s.orchestratorrepo.SelectServer(ctx, region)
	if err != nil {
		return models.Gameserver{}, err
	}
	return Gameserver, nil
}

func (s *Orchestrator) CreateServer(ctx context.Context, region string) error {
	// this will like create a server by talking to the vultr api and being like hey i want a server

	// atleast thats the plan :D

	return nil
}
