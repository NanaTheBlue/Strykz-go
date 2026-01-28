package orchestratorrepo

import (
	"context"
	"time"

	"github.com/nanagoboiler/models"
)

type OrchestratoryRepository interface {
	UpdateHeartBeat(ctx context.Context, serverid string) error
	SelectServer(ctx context.Context, region string) (models.Gameserver, error)
	GetDeadServers(ctx context.Context, cutoff time.Time) ([]models.Gameserver, error)
	GetServersByRegion(ctx context.Context, region string) ([]models.Gameserver, error)
}
