package socialrepo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nanagoboiler/models"
)

type socialRepo struct {
	pool *pgxpool.Pool
}

func NewSocialRepository(pool *pgxpool.Pool) SocialRepository {
	return &socialRepo{pool: pool}
}

func (r *socialRepo) AddFriend(ctx context.Context, notif models.Notification) error {
	a := notif.SenderID
	b := notif.RecipientID

	userID, friendID := a, b
	if a > b {
		userID, friendID = b, a
	}

	_, err := r.pool.Exec(
		ctx,
		"INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)",
		userID,
		friendID,
	)
	return err
}

func (r *socialRepo) RemoveFriend(ctx context.Context, userID string, friendID string) error {
	a := userID
	b := friendID

	user, friend := a, b
	if a > b {
		user, friend = b, a
	}

	_, err := r.pool.Exec(ctx, "DELETE FROM friends Where user_id = $1 AND friend_id = $2", user, friend)
	if err != nil {
		return err
	}
	return nil
}

func (r *socialRepo) IsFriends(ctx context.Context, userID, user2ID string) (bool, error) {
	var exists int

	err := r.pool.QueryRow(ctx, `
        SELECT 1
        FROM friends
        WHERE (user_id = $1 AND friend_id = $2)
           OR (user_id = $2 AND friend_id = $1)
        LIMIT 1;
    `, userID, user2ID).Scan(&exists)

	if err == pgx.ErrNoRows {
		return false, nil
	}

	if err != pgx.ErrNoRows || err != nil {
		return false, err
	}

	return true, nil
}
