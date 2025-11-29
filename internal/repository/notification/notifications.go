package notificationrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nanagoboiler/models"
)

type notificationsRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationsRepository(pool *pgxpool.Pool) NotificationRepository {
	return &notificationsRepo{pool: pool}
}

func (r *notificationsRepo) GetNotifications(ctx context.Context, uuid string) ([]models.Notification, error) {

	rows, err := r.pool.Query(ctx, "SELECT * from notifications WHERE recipient = $1", uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notifications []models.Notification

	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(
			&notification.Sender_id,
			&notification.Recepient_id,
			&notification.Type,
			&notification.Data,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning notification row: %w", err)
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification rows: %w", err)
	}

	return notifications, nil

}

func (r *notificationsRepo) AddFriend(ctx context.Context, notif models.Notification) error {
	a := notif.Sender_id
	b := notif.Recepient_id

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

func (r *notificationsRepo) RemoveFriend(ctx context.Context, userID string, friendID string) error {
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

//todo add block functionality

func (r *notificationsRepo) IsFriends(ctx context.Context, userID, user2ID string) (bool, error) {
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
