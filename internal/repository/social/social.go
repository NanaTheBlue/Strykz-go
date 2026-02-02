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

func (r *socialRepo) AddReport(ctx context.Context, reportreq models.ReportRequestInput) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	cmd, err := tx.Exec(ctx, `INSERT INTO reports (reporter_id, reportee_id, report_type, reason) VALUES ($1, $2 ,$3 ,$4)
		ON CONFLICT DO NOTHING`, reportreq.ReporterID, reportreq.ReporteeID, reportreq.Type, reportreq.Reason)

	if cmd.RowsAffected() == 0 {
		return errors.New("report failed")
	}

	return tx.Commit(ctx)
}

func (r *socialRepo) AddFriend(ctx context.Context, userID string, friendID string) error {
	// check if theres a friend request if not return

	// insert user into friends table

	// delete friend request
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	a, b := normalizePair(userID, friendID)
	defer tx.Rollback(ctx)
	cmd, err := tx.Exec(ctx, `
		INSERT INTO friends (user_id, friend_id)
		SELECT $1, $2
		WHERE EXISTS (
			SELECT 1
			FROM friend_requests
			WHERE sender_id = $3
			  AND recipient_id = $4
		)
		ON CONFLICT DO NOTHING
	`, a, b, friendID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("no pending friend request")
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM friend_requests
		WHERE sender_id = $1
		  AND recipient_id = $2
	`, friendID, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
func (r *socialRepo) IsMutuallyBlocked(ctx context.Context, userA, userB string) (bool, error) {
	blocked1, err := r.IsBlocked(ctx, userA, userB)
	if err != nil {
		return false, err
	}
	blocked2, err := r.IsBlocked(ctx, userB, userA)
	if err != nil {
		return false, err
	}
	return blocked1 || blocked2, nil
}

func (r *socialRepo) AddPartyInvite(ctx context.Context, req models.PartyInviteRequest) (bool, error) {
	cmd, err := r.pool.Exec(ctx, `
		INSERT INTO party_invites (party_id, inviter_id, invitee_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, req.PartyID, req.SenderID, req.RecipientID)
	if err != nil {
		return false, err
	}

	return cmd.RowsAffected() == 1, nil
}

func (r *socialRepo) AddPartyMember(ctx context.Context, req models.PartyInviteRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cmdTag, err := tx.Exec(ctx, `
		WITH deleted_invite AS (
			DELETE FROM party_invites
			WHERE party_id = $1
			  AND invitee_id = $2
			RETURNING party_id
		)
		INSERT INTO party_members (party_id, user_id)
		SELECT party_id, $2
		FROM deleted_invite;
	`, req.PartyID, req.RecipientID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("party invite not found")
	}

	return tx.Commit(ctx)
}
func (r *socialRepo) IsBlocked(ctx context.Context, userID, otherID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM blocks
			WHERE blocker_id = $1 AND blocked_id = $2
		)
	`, userID, otherID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
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

func (r *socialRepo) CreateParty(ctx context.Context, leaderID string) (string, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var partyID string
	err = tx.QueryRow(ctx, `
        INSERT INTO parties (leader_id) VALUES ($1) RETURNING id
    `, leaderID).Scan(&partyID)
	if err != nil {
		return "", err
	}

	return partyID, tx.Commit(ctx)
}

func (r *socialRepo) CheckPartyLeader(ctx context.Context, partyID string) (string, error) {
	var leaderID string
	err := r.pool.QueryRow(ctx, "SELECT leader_id from parties WHERE id =$1", partyID).Scan(&leaderID)
	if err != nil {
		return "", err
	}
	return leaderID, nil
}

func (r *socialRepo) DeleteFriendRequest(ctx context.Context, senderID string, recipientID string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM friend_requests WHERE sender_id = $1 AND recipient_id = $2", senderID, recipientID)
	if err != nil {
		return err
	}
	return nil
}
