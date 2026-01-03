package social

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nanagoboiler/internal/bootstrap"
	notificationrepo "github.com/nanagoboiler/internal/repository/notification"
	redisrepo "github.com/nanagoboiler/internal/repository/redis"
	socialrepo "github.com/nanagoboiler/internal/repository/social"
	"github.com/nanagoboiler/internal/services/notifications"
	"github.com/nanagoboiler/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var (
	testPool  *pgxpool.Pool
	testRedis *redis.Client
)

func TestMain(m *testing.M) {

	fmt.Println("TEST_POSTGRES_URL:", os.Getenv("TEST_POSTGRES_URL"))
	fmt.Println("TEST_REDIS_ADDRESS:", os.Getenv("TEST_REDIS_ADDRESS"))

	envPath := filepath.Join("..", "..", "..", ".env")
	fmt.Println("Loading .env from:", envPath)
	err := godotenv.Load(envPath)
	if err != nil {
		panic("Error loading .env file")
	}

	ctx := context.Background()

	testPool, err = pgxpool.New(ctx, os.Getenv("TEST_POSTGRES_URL"))
	if err != nil {
		panic(err)
	}

	testRedis, err = bootstrap.NewRedisInstance(
		ctx,
		os.Getenv("TEST_REDIS_ADDRESS"),
		os.Getenv("TEST_REDIS_PASSWORD"),
	)
	if err != nil {
		panic(err)
	}
	code := m.Run()

	testPool.Close()
	testRedis.Close()

	os.Exit(code)
}

func setupService(pool *pgxpool.Pool, redisClient *redis.Client) (*socialService, notifications.Service) {
	hub := notifications.NewHub()

	redisStore := redisrepo.NewRedisInstance(redisClient)

	socialRepo := socialrepo.NewSocialRepository(pool)
	notifRepo := notificationrepo.NewNotificationsRepository(pool)

	notifService := notifications.NewnotificationsService(hub, redisStore, notifRepo)

	service := &socialService{
		socialrepo:          socialRepo,
		notificationservice: notifService,
	}

	return service, notifService
}

func createTestUser(t *testing.T, pool *pgxpool.Pool, username string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userID string
	err := pool.QueryRow(ctx,
		`INSERT INTO users (username, email,hashed_password)
		 VALUES ($1,$2,$3)
		 RETURNING id`,
		username,
		username+"@test.com", "TEST",
	).Scan(&userID)

	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(
			ctx,
			`DELETE FROM users WHERE id=$1`,
			userID,
		)
	})
	return userID
}
func TestSendFriendRequest_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	service, _ := setupService(testPool, testRedis)

	user1 := createTestUser(t, testPool, "user1")
	user2 := createTestUser(t, testPool, "user2")

	sub := testRedis.Subscribe(ctx, "notifications")
	defer sub.Close()
	_, err := sub.Receive(ctx)
	require.NoError(t, err)
	ch := sub.Channel()

	err = service.SendFriendRequest(
		ctx,
		models.FriendRequestInput{
			SenderID:    user1,
			RecipientID: user2,
		},
	)
	require.NoError(t, err)
	var exists bool
	err = testPool.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1 FROM friend_requests
			WHERE sender_id=$1 AND receiver_id=$2
		)`,
		user1, user2,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)

	select {
	case msg := <-ch:
		require.Contains(t, msg.Payload, user2)
	case <-time.After(time.Second):
		t.Fatal("did not receive Redis notification")
	}
}

func TestBlockUser_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s, _ := setupService(testPool, testRedis)

	user1 := createTestUser(t, testPool, "user1")
	user2 := createTestUser(t, testPool, "user2")
	err := s.BlockUser(ctx, models.BlockRequest{
		BlockerID: user1,
		BlockedID: user1,
	})
	require.Error(t, err)

	err = s.BlockUser(ctx, models.BlockRequest{
		BlockerID: user1,
		BlockedID: user2,
	})
	require.NoError(t, err)

	var exists bool

	err = testPool.QueryRow(
		ctx, `SELECT EXISTS (
				SELECT 1 FROM blocks
				WHERE blocker_id =$1 AND blocked_id =$2 )`,
		user1, user2,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)

}

func TestRejectNotification_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, s := setupService(testPool, testRedis)
	user1 := createTestUser(t, testPool, "user1")
	user2 := createTestUser(t, testPool, "user2")

	notifID, err := s.CreateNoPublishNotification(ctx, models.Notification{
		SenderID:    user1,
		RecipientID: user2,
		Type:        "Test",
		Status:      "Pending",
	})

	require.NoError(t, err)

	var exists bool

	err = testPool.QueryRow(
		ctx, `SELECT EXISTS (
				SELECT 1 FROM notifications
				WHERE id =$1 )`, notifID,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)

	err = s.DeleteNotification(ctx, notifID)
	require.NoError(t, err)

	var exist bool
	err = testPool.QueryRow(
		ctx, `SELECT EXISTS (
				SELECT 1 FROM notifications
				WHERE id =$1 )`, notifID,
	).Scan(&exist)
	require.NoError(t, err)
	require.False(t, exist)

}
