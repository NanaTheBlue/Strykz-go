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
		//Data
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
