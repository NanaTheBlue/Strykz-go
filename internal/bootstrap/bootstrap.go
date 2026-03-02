package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
)

func ensureDatabase(ctx context.Context, pool *pgxpool.Pool) error {

	scriptPath := filepath.Join("scripts", "databasebuild.sql")

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

	err = retryWithBackoff(ctx, 5, 1*time.Second, func() error {
		return pool.Ping(ctx)
	})

	if err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	err = retryWithBackoff(ctx, 5, 1*time.Second, func() error {
		return ensureDatabase(ctx, pool)
	})
	if err != nil {
		return nil, fmt.Errorf("ensure database: %w", err)
	}

	return pool, nil

}

func NewRedisInstance(ctx context.Context, address string, password string) (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0, // Use default DB
		Protocol: 2, // Connection protocol
	})
	//ping the redis client to ensure it works properly

	err := retryWithBackoff(ctx, 5, 1*time.Second, func() error {
		return client.Ping(ctx).Err()
	})
	if err != nil {
		return nil, err
	}

	return client, nil

}
