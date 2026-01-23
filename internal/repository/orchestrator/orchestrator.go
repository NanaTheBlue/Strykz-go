package orchestratorrepo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type orchastratorRepo struct {
	pool *pgxpool.Pool
}

func NewOrchestratorRepository(pool *pgxpool.Pool) OrchestratoryRepository {
	return &orchastratorRepo{pool: pool}
}

func (r *orchastratorRepo) UpdateHeartBeat(serverid string, ctx context.Context) error {
	currentTime := time.Now()

	_, err := r.pool.Exec(ctx, "UPDATE game_servers SET last_heartbeat = $1  WHERE id = $2 ", currentTime, serverid)
	if err != nil {
		return err
	}
	return nil
}
