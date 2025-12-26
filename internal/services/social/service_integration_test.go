package social

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	notificationrepo "github.com/nanagoboiler/internal/repository/notification"
	"github.com/nanagoboiler/internal/repository/redis"
	socialrepo "github.com/nanagoboiler/internal/repository/social"
	"github.com/nanagoboiler/internal/services/notifications"
)

func setupService(pool *pgxpool.Pool) (*socialService, notifications.Service) {
	hub := notifications.NewHub()

	redisClient := redis.InitRedis()
	redisStore := redis.NewRedisInstance(redisClient)

	socialRepo := socialrepo.NewSocialRepository(pool)
	notifRepo := notificationrepo.NewNotificationsRepository(pool)

	notifService := notifications.NewnotificationsService(hub, redisStore, notifRepo)

	service := &socialService{
		socialrepo:          socialRepo,
		notificationservice: notifService,
	}

	return service, notifService
}
func TestSendFriendRequest_Integration(t *testing.T) {

}
