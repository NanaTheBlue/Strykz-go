package txutil

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func WithTx(
	ctx context.Context,
	db Beginner,
	fn func(tx pgx.Tx) error,
) error {
	if db == nil {
		return errors.New("database beginner is nil")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
