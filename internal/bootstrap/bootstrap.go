package bootstrap

import (
	"context"

	"fmt"

	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/joho/godotenv/autoload"
)

func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {

	postgresURL := os.Getenv("POSTGRES_URL")

	println(postgresURL)

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil

}

func newRedisInstance()
