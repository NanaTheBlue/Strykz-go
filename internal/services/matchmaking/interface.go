package matchmaking

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Service interface {
	InQue(ctx context.Context, player *models.Player) error
	DeQue(ctx context.Context, mode string, region string, count int) ([]*models.Player, error)
	DeQuePlayer(ctx context.Context, mode string, region string, playerID string) error
}
