package matchmaking

import (
	"context"

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
