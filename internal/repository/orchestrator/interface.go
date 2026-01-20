package orchestratorrepo

import "context"

type OrchestratoryRepository interface {
	UpdateHeartBeat(serverid string, ctx context.Context) error
}
