package authrepo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type tokensRepo struct {
	pool *pgxpool.Pool
}

func NewTokensRepository(pool *pgxpool.Pool) TokensRepository {
	return &tokensRepo{pool: pool}
}

func (r *tokensRepo) AddRefresh(ctx context.Context, jti string, uuid string) error {
	refreshExpiry := time.Now().UTC().Add(30 * 24 * time.Hour)
	_, err := r.pool.Exec(ctx, "UPDATE users SET refresh_token = $1, expires_at = $2 WHERE id = $3;", jti, refreshExpiry, uuid)
	if err != nil {
		return err
	}

	return nil

}
