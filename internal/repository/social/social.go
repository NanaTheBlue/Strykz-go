package socialrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nanagoboiler/models"
)

type socialRepo struct {
	pool *pgxpool.Pool
}

func NewSocialRepository(pool *pgxpool.Pool) SocialRepository {
	return &socialRepo{pool: pool}
}

func (r *socialRepo) AddFriend(ctx context.Context, userID string, friendID string) error {

	_, err := r.pool.Exec(
		ctx,
		"INSERT INTO friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
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

func (r *socialRepo) IsBlocked(ctx context.Context, userID string, blockedID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM blocks
            WHERE blocker_id = $1 AND blocked_id = $2
        );
    `, userID, blockedID).Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil

}

func (r *socialRepo) BlockUser(ctx context.Context, blockreq models.BlockRequest) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO blocks (blocker_id, blocked_id) VALUES ($1, $2)",
		blockreq.BlockerID, blockreq.BlockedID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *socialRepo) CreateFriendRequest(ctx context.Context, senderID string, recipientID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO friend_requests (sender_id, recipient_id) 
         VALUES ($1, $2)
         ON CONFLICT DO NOTHING`,
		senderID,
		recipientID,
	)
	return err
}
func (r *socialRepo) DeleteFriendRequest(ctx context.Context, senderID string, recipientID string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM friend_requests WHERE sender_id = $1 AND recipient_id = $2", senderID, recipientID)
	if err != nil {
		return err
	}
	return nil
}

func (r *socialRepo) IsFriends(ctx context.Context, userID, user2ID string) (bool, error) {
	var exists bool

	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM friends
            WHERE (user_id = $1 AND friend_id = $2)
        );
    `, userID, user2ID).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}
