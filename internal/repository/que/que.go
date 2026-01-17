package querepo

import "github.com/jackc/pgx/v5/pgxpool"

type queRepo struct {
	pool *pgxpool.Pool
}

func NewQueRepository(pool *pgxpool.Pool) QueRepository {
	return &queRepo{pool: pool}
}

func (r *queRepo) CheckBan(player string) (bool, error) {

	return false, nil
}
