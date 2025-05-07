package db

import (
	"context"

	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Pool *pgxpool.Pool
	once sync.Once
)

func InitDB() {

	once.Do(func() {
		var err error
		Pool, err = pgxpool.New(context.Background(), "postgres://postgres:8575@localhost:5432/strykz_database?sslmode=disable")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}
	})
}

func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}
