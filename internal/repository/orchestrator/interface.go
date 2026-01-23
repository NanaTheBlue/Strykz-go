package orchestratorrepo

import (
	"context"

	"github.com/nanagoboiler/models"
)

type OrchestratoryRepository interface {
	UpdateHeartBeat(serverid string, ctx context.Context) error
	SelectServer(ctx context.Context, region string) (models.Gameserver, error)
}
