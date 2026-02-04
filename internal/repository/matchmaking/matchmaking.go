package matchmakingrepo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/nanagoboiler/internal/repository/db"
	"github.com/nanagoboiler/models"
)

type matchmakingRepo struct {
	db db.DB
}

func NewMatchmakingRepository(db db.DB) MatchmakingRepository {
	return &matchmakingRepo{db: db}
}

func (r *matchmakingRepo) CreateMatch(ctx context.Context, serverid string) error {
	var matchID string

	err := r.db.QueryRow(ctx, `
        INSERT INTO matches (server_id)
        VALUES ($1)
        RETURNING id
    `, serverid).Scan(&matchID)
	if err != nil {
		return err
	}
	return nil
}

func (r *matchmakingRepo) InsertPlayers(ctx context.Context, players []models.Player, matchid string) error {
	batch := &pgx.Batch{}

	for _, p := range players {
		batch.Queue(`
            INSERT INTO match_players (match_id, steam_id, status)
            VALUES ($1, $2, 'unconnected')
        `, matchid, p.Player_steamid)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range players {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *matchmakingRepo) UpdatePlayer(ctx context.Context, player models.Player, matchid string, status string) error {

	_, err := r.db.Exec(ctx, `UPDATE match_players SET status=$1 WHERE match_id= $2 AND steam_id =$3 `, status, matchid, player.Player_steamid)
	if err != nil {
		return err
	}
	return nil
}
