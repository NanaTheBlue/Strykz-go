package matchmaking

import (
	"context"

	"fmt"

	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/models"
)

type matchmakingService struct {
	RedisRepo redis.Store
}

func NewMatchmakingService(redisRepo redis.Store) Service {
	return &matchmakingService{RedisRepo: redisRepo}
}

func (s *matchmakingService) InQue(ctx context.Context, player *models.Player) error {

	err := s.RedisRepo.Que(ctx, "1v1", "us", player)
	if err != nil {
		return err
	}

	return nil
}

func (s *matchmakingService) DeQue(ctx context.Context, mode string, region string, count int) ([]*models.Player, error) {

	players, err := s.RedisRepo.DeQue(ctx, mode, region, count)
	if err != nil {
		return []*models.Player{}, err
	}

	return players, nil
}

func (s *matchmakingService) DeQuePlayer(ctx context.Context, mode string, region string, playerID string) error {
	err := s.RedisRepo.DeQuePlayer(ctx, mode, region, playerID)

	if err != nil {
		return err
	}

	return nil
}

func (s *matchmakingService) StartMatchMaking(ctx context.Context) {
	regions := []string{"us"}
	modes := []string{"1v1"}

	for {
		for _, mode := range modes {
			for _, region := range regions {
				queueKey := fmt.Sprintf("queue:%s:%s", mode, region)

				matchCandidates, err := s.DeQue(ctx, queueKey, region, 2)
				if err != nil {
					log.Printf("Error reading from queue %s: %v", queueKey, err)
					continue
				}

				if matchCandidates == nil {
					log.Printf("No candidates found")
					time.Sleep(5 * time.Second)
					continue
				}

				player1 := matchCandidates[0]
				player2 := matchCandidates[1]

				matchID := uuid.New().String()
				log.Printf("Creating match %s between %s and %s", matchID, player1.Player_id, player2.Player_id)

			}

		}

	}

}
