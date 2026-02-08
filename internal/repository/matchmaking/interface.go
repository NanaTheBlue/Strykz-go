package matchmakingrepo

import (
	"context"

	"github.com/nanagoboiler/models"
)

type MatchmakingRepository interface {
	InsertPlayers(ctx context.Context, players []*models.Player, matchid string) error
	CreateMatch(ctx context.Context, serverid string) (string, error)
	UpdatePlayer(ctx context.Context, player models.Player, matchid string, status string) error
}
