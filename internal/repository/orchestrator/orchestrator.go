package orchestratorrepo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nanagoboiler/internal/repository/db"
	"github.com/nanagoboiler/models"
)

type orchestratorRepo struct {
	db db.DB
}

func NewOrchestratorRepository(db db.DB) OrchestratoryRepository {
	return &orchestratorRepo{db: db}
}

func (r *orchestratorRepo) UpdateHeartBeat(ctx context.Context, serverid string) error {
	currentTime := time.Now()

	_, err := r.db.Exec(ctx, "UPDATE game_servers SET last_heartbeat = $1  WHERE id = $2 ", currentTime, serverid)
	if err != nil {
		return err
	}
	return nil
}

func (r *orchestratorRepo) UpdateServer(ctx context.Context, id string, status models.ServerStatus) error {

	return nil
}

func (r *orchestratorRepo) InsertServer(ctx context.Context, server models.Gameserver) error {
	_, err := r.db.Exec(ctx, "INSERT INTO game_servers (id, region, status) VALUES ($1, $2, $3)", server.ID, server.Region, server.Status)
	if err != nil {
		return err
	}

	return nil
}
func (r *orchestratorRepo) AcquireReadyServer(ctx context.Context, region string) (*models.Gameserver, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE game_servers
		SET status = 'BUSY'
		WHERE id = (
			SELECT id
			FROM game_servers
			WHERE region = $1 AND status = 'READY'
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, region, status, last_heartbeat, created_at
	`, region)

	var s models.Gameserver
	if err := row.Scan(
		&s.ID,
		&s.Region,
		&s.Status,
		&s.LastHeartbeat,
		&s.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (r *orchestratorRepo) DeleteServer(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM game_servers WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (r *orchestratorRepo) GetDeadServers(ctx context.Context, cutoff time.Time) ([]models.Gameserver, error) {

	// cutoff should be Time.Now() - 1 minute
	rows, err := r.db.Query(ctx, "SELECT id, region, status, last_heartbeat, created_at FROM game_servers WHERE last_heartbeat < $1", cutoff)
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

func (r *orchestratorRepo) GetServersByRegion(ctx context.Context, region string) ([]models.Gameserver, error) {

	rows, err := r.db.Query(ctx, "SELECT * FROM game_servers WHERE region = $1", region)
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

func (r *orchestratorRepo) WithTx(tx pgx.Tx) OrchestratoryRepository {
	return &orchestratorRepo{
		db: tx,
	}
}
