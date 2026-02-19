package matchmakingrepo

import (
	"context"
	"time"

	"github.com/nanagoboiler/models"
)

type MatchmakingRepository interface {
	InsertPlayers(ctx context.Context, players []*models.Player, matchid string) error
	CreateMatch(ctx context.Context, deadline time.Time, region string) (string, error)
	UpdatePlayer(ctx context.Context, player models.Player, matchid string, status string) error
	AreAllPlayersAccepted(ctx context.Context, matchID string) (bool, error)
	UpdateMatchStatus(ctx context.Context, matchID string, status models.MatchStatus) error
	AssignServerToMatch(ctx context.Context, matchID string, serverID string) error
	GetMatch(ctx context.Context, matchID string) (models.Match, error)
	GetMatchPlayers(ctx context.Context, matchID string) ([]models.Player, error)
	GetMatchesByStatus(ctx context.Context, status models.MatchStatus) ([]models.Match, error)
}
