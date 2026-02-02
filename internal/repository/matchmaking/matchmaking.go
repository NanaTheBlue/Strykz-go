package matchmakingrepo

import (
	"context"

	"github.com/nanagoboiler/internal/repository/db"
	"github.com/nanagoboiler/models"
)

type matchmakingRepo struct {
	db db.DB
}

func NewMatchmakingRepository(db db.DB) MatchmakingRepository {
	return &matchmakingRepo{db: db}
}

func (r *matchmakingRepo) CreateMatch(ctx context.Context, players []models.Player) error {

	_, err := r.db.Exec(ctx, "INSERT INTO matches ()")
	if err != nil {
		return err
	}
	return nil
}
