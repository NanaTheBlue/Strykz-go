package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	pb "github.com/nanagoboiler/gen"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	"github.com/nanagoboiler/models"
	"github.com/vultr/govultr/v3"
)

type Orchestrator struct {
	orchestratorrepo orchestratorrepo.OrchestratoryRepository
	vultrclient      *govultr.Client
	streams          map[string]pb.SidecarService_ConnectServer
	mu               sync.RWMutex
}

func NewOrchestrator(orchestratorrepo orchestratorrepo.OrchestratoryRepository, vultrclient *govultr.Client) Service {
	return &Orchestrator{
		orchestratorrepo: orchestratorrepo,
		vultrclient:      vultrclient,
		streams:          make(map[string]pb.SidecarService_ConnectServer),
	}
}

func (s *Orchestrator) RegisterStream(serverID string, stream pb.SidecarService_ConnectServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams[serverID] = stream
	log.Printf("registered stream for server %s", serverID)
}

func (s *Orchestrator) UnregisterStream(serverID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.streams, serverID)
	log.Printf("unregistered stream for server %s", serverID)
}

func (s *Orchestrator) GetStream(serverID string) pb.SidecarService_ConnectServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.streams[serverID]
}
func (s *Orchestrator) ReloadWhitelist(serverID string, steamIDs []string) error {
	// the game server would clear its whitelist after the match
	stream := s.GetStream(serverID)
	if stream == nil {
		return fmt.Errorf("no stream for server %s", serverID)
	}

	cmd := &pb.BackendCommand{
		Payload: &pb.BackendCommand_ReloadWhitelist{
			ReloadWhitelist: &pb.ReloadWhitelist{
				SteamIds: steamIDs,
			},
		},
	}

	if err := stream.Send(cmd); err != nil {
		log.Printf("failed to send reload whitelist to server %s: %v", serverID, err)
		return err
	}

	log.Printf("reload whitelist sent to server %s", serverID)
	return nil
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

func (s *Orchestrator) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	err := s.orchestratorrepo.UpdateServer(ctx, serverID, status)
	if err != nil {
		return err
	}
	return nil
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
	if instance.ID == "" {
		return "", errors.New("Instance ID Is Blank")
	}

	server := models.Gameserver{
		ID:     instance.ID,
		Region: region,
		Status: "Creating",
	}

	err = s.orchestratorrepo.InsertServer(ctx, server)
	if err != nil {
		return "", err
	}

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
