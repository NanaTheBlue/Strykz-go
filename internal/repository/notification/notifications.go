package notificationrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nanagoboiler/models"
)

type notificationsRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationsRepository(pool *pgxpool.Pool) NotificationRepository {
	return &notificationsRepo{pool: pool}
}

func (r *notificationsRepo) GetNotification(ctx context.Context, notificationID string) (models.Notification, error) {
	var notification models.Notification
	err := r.pool.QueryRow(ctx, "SELECT sender_id, addressee_id, notification_type, data, created_at FROM notifications WHERE id =$1 ", notificationID).Scan(&notification.SenderID,
		&notification.RecipientID,
		&notification.Type,
		&notification.Data,
		&notification.CreatedAt)
	if err != nil {
		return models.Notification{}, err
	}

	return notification, nil
}
func (r *notificationsRepo) SendNotification(ctx context.Context, notif models.Notification) (string, error) {
	var id string

	err := r.pool.QueryRow(ctx,
		`INSERT INTO notifications (sender_id, addressee_id, notification_status, notification_type)
     VALUES ($1, $2, $3, $4)
	 RETURNING ID`,
		notif.SenderID, notif.RecipientID, notif.Status, notif.Type,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *notificationsRepo) GetNotifications(ctx context.Context, uuid string) ([]models.Notification, error) {

	rows, err := r.pool.Query(ctx, "SELECT * FROM notifications WHERE recipient = $1", uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notifications []models.Notification

	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(
			&notification.ID,
			&notification.SenderID,
			&notification.RecipientID,
			&notification.Type,
			&notification.Data,
			&notification.CreatedAt,
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

func (r *notificationsRepo) DeleteNotification(ctx context.Context, notificationID string) error {

	_, err := r.pool.Exec(ctx, "DELETE FROM notifications where id = $1", notificationID)
	if err != nil {
		return err
	}
	return nil

}
