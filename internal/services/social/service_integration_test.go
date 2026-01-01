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

	var userID string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (username, email,hashed_password)
		 VALUES ($1, $2,$3)
		 RETURNING id`,
		username,
		username+"@test.com", "TEST",
	).Scan(&userID)

	require.NoError(t, err)
	return userID
}
func TestSendFriendRequest_Integration(t *testing.T) {
	ctx := context.Background()
	service, _ := setupService(testPool, testRedis)

	user1 := createTestUser(t, testPool, "user1")
	user2 := createTestUser(t, testPool, "user2")

	sub := testRedis.Subscribe(ctx, "notifications")
	defer sub.Close()
	_, err := sub.Receive(ctx)
	require.NoError(t, err)
	ch := sub.Channel()

	err = service.SendFriendRequest(
		context.Background(),
		models.FriendRequestInput{
			SenderID:    user1,
			RecipientID: user2,
		},
	)
	require.NoError(t, err)
	var exists bool
	err = testPool.QueryRow(
		context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM friend_requests
			WHERE sender_id=$1 AND receiver_id=$2
		)`,
		user1, user2,
	).Scan(&exists)

	select {
	case msg := <-ch:
		require.Contains(t, msg.Payload, user2)
	case <-time.After(time.Second):
		t.Fatal("did not receive Redis notification")
	}
}
