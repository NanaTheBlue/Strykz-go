package matchmaking

import (
	"context"

	"fmt"

	"log"
	"time"

	"github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/models"
)

// will prob need to add a matchmaking repo
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

func (s *matchmakingService) StartMatchMaking(ctx context.Context, mode string) {

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.QueReader(ctx, mode)
		}
	}

}

// need to write tests for this when i wake up
func (s *matchmakingService) QueReader(ctx context.Context, mode string) {
	regions := []string{"us"}
	modes := []string{"1v1"}

	for _, mode := range modes {
		for _, region := range regions {
			queueKey := fmt.Sprintf("queue:%s:%s", mode, region)

			matchCandidates, err := s.RedisRepo.DeQue(ctx, queueKey, region, 2)
			if err != nil {
				log.Printf("Error reading from queue %s: %v", queueKey, err)
				continue
			}

			if len(matchCandidates) < 2 {

				continue
			}

			go s.CreateMatch(ctx, matchCandidates)

		}

	}

}

func (s *matchmakingService) CreateMatch(ctx context.Context, matchCanidates []*models.Player) {

	// put the match into the database obviously then
	// assign them a server
	// obviously this will get more complicated whe we are dealing with 5v5 mode
	// notify players

}
