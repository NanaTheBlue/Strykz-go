package bootstrap

import (
	"context"
	"os"
	"path/filepath"

	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/joho/godotenv/autoload"
)

func ensureDatabase(ctx context.Context, pool *pgxpool.Pool) error {

	scriptPath := filepath.Join("..", "scripts", "databasebuild.sql")

	sqlBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read SQL script: %w", err)
	}

	_, err = pool.Exec(ctx, string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute SQL script: %w", err)
	}

	fmt.Println("Database schema applied successfully!")
	return nil
}

func NewPostgresPool(ctx context.Context, postgresURL string) (*pgxpool.Pool, error) {

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	err = ensureDatabase(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("ensure database: %w", err)
	}

	return pool, nil

}

func newRedisInstance() {

}
