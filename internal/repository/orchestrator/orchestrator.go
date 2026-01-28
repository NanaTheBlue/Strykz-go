package orchestratorrepo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nanagoboiler/models"
)

type orchestratorRepo struct {
	pool *pgxpool.Pool
}

func NewOrchestratorRepository(pool *pgxpool.Pool) OrchestratoryRepository {
	return &orchestratorRepo{pool: pool}
}

func (r *orchestratorRepo) UpdateHeartBeat(ctx context.Context, serverid string) error {
	currentTime := time.Now()

	_, err := r.pool.Exec(ctx, "UPDATE game_servers SET last_heartbeat = $1  WHERE id = $2 ", currentTime, serverid)
	if err != nil {
		return err
	}
	return nil
}

func (r *orchestratorRepo) GetDeadServers(ctx context.Context, cutoff time.Time) ([]models.Gameserver, error) {

 rows,err : = rows, err := r.pool.Query(ctx, "SELECT * FROM game_servers WHERE regions = $1", cutoff)

}

func (r *orchestratorRepo) GetServersByRegion(ctx context.Context, region string) ([]models.Gameserver, error) {

	rows, err := r.pool.Query(ctx, "SELECT * FROM game_servers WHERE regions = $1", region)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []models.Gameserver

	for rows.Next() {
		var s models.Gameserver
		if err := rows.Scan(
			&s.ID,
			&s.Region,
			&s.Status,
			&s.LastHeartbeat,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, rows.Err()

}

func (r *orchestratorRepo) SelectServer(ctx context.Context, region string) (models.Gameserver, error) {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Gameserver{}, err
	}
	defer tx.Rollback(ctx)

	var Gameserver models.Gameserver

	err = tx.QueryRow(ctx, "SELECT id,region,status,last_heartbeat FROM game_servers WHERE region = $1 AND status = $2 FOR UPDATE SKIP LOCKED", region, "ready").Scan(
		&Gameserver.ID,
		&Gameserver.Region,
		&Gameserver.Status,
		&Gameserver.LastHeartbeat,
	)
	if err != nil {
		return models.Gameserver{}, err
	}

	_, err = tx.Exec(ctx, "UPDATE game_servers SET status = $1, updated_at = $2 WHERE id = $3 ", "used", time.Now(), Gameserver.ID)
	if err != nil {
		return models.Gameserver{}, err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return models.Gameserver{}, err
	}
	Gameserver.Status = "used"
	return Gameserver, nil
}
