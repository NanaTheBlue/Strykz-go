package matchmaking

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Service interface {
	InQue(ctx context.Context, player *models.Player) error
	QueReader(ctx context.Context, mode string)
	StartMatchMaking(ctx context.Context, mode string)
	CreateMatch(ctx context.Context, matchCanidates []*models.Player)
}
