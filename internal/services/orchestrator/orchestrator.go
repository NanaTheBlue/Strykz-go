package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	pb "github.com/nanagoboiler/gen"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	gameserverconfig "github.com/nanagoboiler/internal/services/config"
	"github.com/nanagoboiler/models"
)

type Orchestrator struct {
	orchestratorrepo orchestratorrepo.OrchestratoryRepository
	ec2client        *ec2.Client
	streams          map[string]pb.SidecarService_ConnectServer
	mu               sync.RWMutex
	cfg              gameserverconfig.Config
}

func NewOrchestrator(orchestratorrepo orchestratorrepo.OrchestratoryRepository, ec2client *ec2.Client, cfg gameserverconfig.Config) Service {
	return &Orchestrator{
		orchestratorrepo: orchestratorrepo,
		ec2client:        ec2client,
		streams:          make(map[string]pb.SidecarService_ConnectServer),
		cfg:              cfg,
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
		s.UnregisterStream(serverID)
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

	instanceOptions := &ec2.RunInstancesInput{
		ImageId:      aws.String(s.cfg.AMI),
		InstanceType: types.InstanceType(s.cfg.InstanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		SubnetId:     aws.String(s.cfg.SubnetID),
		SecurityGroupIds: []string{
			s.cfg.SecurityGroup,
		},
	}

	instance, err := s.ec2client.RunInstances(ctx, instanceOptions)
	if err != nil {
		return "", err
	}
	if len(instance.Instances) == 0 {
		return "", errors.New("no instances created")
	}

	instanceID := aws.ToString(instance.Instances[0].InstanceId)
	if instanceID == "" {
		return "", errors.New("instance id is empty")
	}

	server := models.Gameserver{
		ID:     instanceID,
		IP:     "",
		Region: region,
		Status: "Creating",
	}

	err = s.orchestratorrepo.InsertServer(ctx, server)
	if err != nil {
		return "", err
	}

	return instanceID, nil
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
