package matchmakingrepo

import (
	"context"
	"time"

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

func (r *matchmakingRepo) CreateMatch(ctx context.Context, deadline time.Time) (string, error) {
	var matchID string

	err := r.db.QueryRow(ctx, `
        INSERT INTO matches (accept_deadline)
        VALUES ($1)
        RETURNING id
    `, deadline).Scan(&matchID)
	if err != nil {
		return "", err
	}
	return matchID, nil
}

func (r *matchmakingRepo) GetMatchPlayers(ctx context.Context, matchID string) ([]models.Player, error) {
	rows, err := r.db.Query(ctx, `
		SELECT steam_id, status, joined_at
		FROM match_players
		WHERE match_id = $1
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []models.Player

	for rows.Next() {
		var steamID string
		var status string
		var joinedAt *time.Time

		if err := rows.Scan(&steamID, &status, &joinedAt); err != nil {
			return nil, err
		}

		var joinedUnix int64
		if joinedAt != nil {
			joinedUnix = joinedAt.Unix()
		}

		player := models.Player{
			Player_id:      steamID,
			Player_steamid: steamID,
			JoinedAt:       joinedUnix,
			Status:         status,
		}

		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}

func (r *matchmakingRepo) GetMatch(ctx context.Context, matchID string) (models.Match, error) {
	var match models.Match
	err := r.db.QueryRow(ctx, "SELECT id, server_id, started_at, accept_deadline, status, ended_at FROM matches WHERE id = $1", matchID).Scan(&match.ID,
		&match.ServerID,
		&match.StartedAt,
		&match.AcceptDeadline,
		&match.Status,
		&match.EndedAt)

	if err != nil {
		return models.Match{}, err
	}
	return match, nil
}

func (r *matchmakingRepo) AreAllPlayersAccepted(ctx context.Context, matchID string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM match_players
            WHERE match_id = $1
              AND status != 'accepted'
        )
    `, matchID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return !exists, nil

}

func (r *matchmakingRepo) AssignServerToMatch(ctx context.Context, matchID string, serverID string) error {
	_, err := r.db.Exec(ctx, "UPDATE matches SET server_id = $2 WHERE id = $1", matchID, serverID)
	if err != nil {
		return err
	}
	return nil
}

func (r *matchmakingRepo) UpdateMatchStatus(ctx context.Context, matchID string, status string) error {
	_, err := r.db.Exec(ctx, "UPDATE matches SET status = $2 WHERE id = $1 ", matchID, status)
	if err != nil {
		return err
	}
	return nil
}

func (r *matchmakingRepo) InsertPlayers(ctx context.Context, players []*models.Player, matchid string) error {
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
