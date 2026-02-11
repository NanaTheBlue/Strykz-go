package orchestrator

import (
	"context"
	"log"

	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	"github.com/nanagoboiler/models"
	"github.com/vultr/govultr/v3"
)

type Orchestrator struct {
	orchestratorrepo orchestratorrepo.OrchestratoryRepository
	vultrclient      *govultr.Client
}

func NewOrchestrator(orchestratorrepo orchestratorrepo.OrchestratoryRepository, vultrclient *govultr.Client) Service {
	return &Orchestrator{
		orchestratorrepo: orchestratorrepo,
		vultrclient:      vultrclient,
	}
}

func (s *Orchestrator) UpdateHeartbeat(ctx context.Context, serverID string) error {

	//TODO: Better Error Handling
	return s.orchestratorrepo.UpdateHeartBeat(ctx, serverID)
}

func (s *Orchestrator) SelectServer(ctx context.Context, region string) (*models.Gameserver, error) {
	//TODO: Better Error Handling honestly i should make this a github issue

	Gameserver, err := s.orchestratorrepo.AcquireReadyServer(ctx, region)
	if err != nil {
		return &models.Gameserver{}, err
	}
	return Gameserver, nil
}

func (s *Orchestrator) CreateServer(ctx context.Context, region string) (string, error) {
	// todo make this more modular rn its just in testing phase so it dont matter
	enableIPv6 := false
	instanceOptions := &govultr.InstanceCreateReq{
		Label:      "awesome-go-app",
		Hostname:   "awesome-go.com",
		Backups:    "enabled",
		EnableIPv6: &enableIPv6,
		OsID:       2284,
		Plan:       "vc2-1c-1gb",
		Region:     region,
	}

	instance, _, err := s.vultrclient.Instance.Create(ctx, instanceOptions)
	if err != nil {
		return "", err
	}

	//todo add the Server to the database

	return instance.ID, nil
}

func (s *Orchestrator) Request(region string) {
	go func() {
		ctx := context.Background()

		ready, err := s.orchestratorrepo.CountReadyServers(ctx, region)
		if err != nil || ready > 0 {
			return
		}

		_, err = s.CreateServer(ctx, region)
		if err != nil {
			log.Println("failed to create server:", err)
			return
		}
	}()
}
