package socialrepo

import (
	"context"
	"errors"

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

	a, b := normalizePair(userID, friendID)

	_, err := r.pool.Exec(ctx, "DELETE FROM friends Where user_id = $1 AND friend_id = $2", a, b)
	if err != nil {
		return err
	}
	return nil
}

func (r *socialRepo) BlockUser(ctx context.Context, blockreq models.BlockRequest) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	cmd, err := tx.Exec(ctx, `
		INSERT INTO blocks (blocker_id, blocked_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, blockreq.BlockerID, blockreq.BlockedID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("user is already blocked")
	}

	a, b := normalizePair(blockreq.BlockerID, blockreq.BlockedID)

	_, err = tx.Exec(ctx, `
		DELETE FROM friends
		WHERE user_id = $1 AND friend_id = $2
	`, a, b)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		DELETE FROM friend_requests
		WHERE (sender_id = $1 AND recipient_id = $2)
   		OR (sender_id = $2 AND recipient_id = $1);`, a, b)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *socialRepo) CreateFriendRequest(ctx context.Context, friendreq models.FriendRequestInput) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	a, b := normalizePair(friendreq.SenderID, friendreq.RecipientID)
	cmd, err := tx.Exec(ctx, `
		INSERT INTO friend_requests (sender_id, recipient_id)
		SELECT $1, $2
		WHERE NOT EXISTS (
    	SELECT 1 FROM friends WHERE user_id = $3 AND friend_id = $4
		)
		AND NOT EXISTS (
    	SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2
		)
		AND NOT EXISTS (
    	SELECT 1 FROM blocks WHERE blocker_id = $2 AND blocked_id = $1
						)
        AND NOT EXISTS (
   		SELECT 1
    	FROM friend_requests
    	WHERE (sender_id = $1 AND recipient_id = $2)
       	OR (sender_id = $2 AND recipient_id = $1)
);
		
	`, friendreq.SenderID, friendreq.RecipientID, a, b)
	if err != nil {
		return err
	}
	// Todo: better error handling
	if cmd.RowsAffected() == 0 {
		return errors.New("friend request cannot be created: already friends or blocked")
	}

	return tx.Commit(ctx)
}

func (r *socialRepo) CreateParty(ctx context.Context, leaderID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var partyID string
	err = tx.QueryRow(ctx, `
        INSERT INTO parties (leader_id) VALUES ($1) RETURNING id
    `, leaderID).Scan(&partyID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO party_members (party_id, user_id) VALUES ($1, $2)
    `, partyID, leaderID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *socialRepo) DeleteFriendRequest(ctx context.Context, senderID string, recipientID string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM friend_requests WHERE sender_id = $1 AND recipient_id = $2", senderID, recipientID)
	if err != nil {
		return err
	}
	return nil
}
