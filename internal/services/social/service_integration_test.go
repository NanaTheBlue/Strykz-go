//go:build integration

package social

import (
	"context"
	"encoding/json"
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func TestMain(m *testing.M) {
	envPath := filepath.Join("..", "..", "..", ".env")
	fmt.Println("Loading .env from:", envPath)
	_ = godotenv.Load(envPath)

	postgresURL := firstNonEmpty(os.Getenv("TEST_POSTGRES_URL"), os.Getenv("POSTGRES_URL"))
	redisAddr := firstNonEmpty(os.Getenv("TEST_REDIS_ADDRESS"), os.Getenv("TEST_REDIS_ADDR"), os.Getenv("REDIS_ADDRESS"))
	redisPass := firstNonEmpty(os.Getenv("TEST_REDIS_PASSWORD"), os.Getenv("REDIS_PASSWORD"))

	fmt.Println("TEST_POSTGRES_URL:", postgresURL)
	fmt.Println("TEST_REDIS_ADDRESS:", redisAddr)

	ctx := context.Background()

	var err error
	testPool, err = pgxpool.New(ctx, postgresURL)
	if err != nil {
		panic(err)
	}

	testRedis, err = bootstrap.NewRedisInstance(
		ctx,
		redisAddr,
		redisPass,
	)
	if err != nil {
		panic(err)
	}
	code := m.Run()

	testPool.Close()
	_ = testRedis.Close()

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
		store:               redisStore,
		db:                  pool,
	}

	return service, notifService
}

func createTestUser(t *testing.T, pool *pgxpool.Pool, username string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = pool.Exec(ctx, `DELETE FROM parties WHERE leader_id IN (SELECT id FROM users WHERE username=$1)`, username)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE username=$1`, username)

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
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM party_invites WHERE inviter_id=$1 OR invitee_id=$1`, userID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM party_members WHERE user_id=$1`, userID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM parties WHERE leader_id=$1`, userID)
		_, _ = pool.Exec(
			cleanupCtx,
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
	defer func() { _ = sub.Close() }()
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
			WHERE sender_id=$1 AND recipient_id=$2
		)`,
		user1, user2,
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)

	select {
	case msg := <-ch:
		require.Contains(t, msg.Payload, user2)
	case <-ctx.Done():
		t.Fatal("did not receive Redis notification")
	}
}

func TestBlockUser_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s, _ := setupService(testPool, testRedis)

	user1 := createTestUser(t, testPool, "user1")
	user2 := createTestUser(t, testPool, "user2")

	err := s.BlockUser(ctx, user1, user2)
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
	social, notification := setupService(testPool, testRedis)
	user1 := createTestUser(t, testPool, "user1")
	user2 := createTestUser(t, testPool, "user2")

	notifID, err := notification.CreateNoPublishNotification(ctx, models.Notification{
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

	err = social.RejectNotification(ctx, notifID)
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

func TestPartyInvite_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	s, _ := setupService(testPool, testRedis)
	user1 := createTestUser(t, testPool, "user1")
	user2 := createTestUser(t, testPool, "user2")

	err := s.socialrepo.CreateFriendRequest(ctx, models.FriendRequestInput{
		SenderID:    user1,
		RecipientID: user2,
	})
	require.NoError(t, err)
	err = s.socialrepo.AddFriend(ctx, user2, user1)
	require.NoError(t, err)

	sub := testRedis.Subscribe(ctx, "notifications")
	defer func() { _ = sub.Close() }()
	_, err = sub.Receive(ctx)
	require.NoError(t, err)
	ch := sub.Channel()

	partyID, err := s.CreateParty(ctx, user1)
	require.NoError(t, err)

	err = s.PartyInvite(ctx, models.PartyInviteRequest{
		PartyID:     partyID,
		SenderID:    user1,
		RecipientID: user2,
	})
	require.NoError(t, err)

	select {
	case msg := <-ch:
		var notif models.Notification
		err := json.Unmarshal([]byte(msg.Payload), &notif)
		require.NoError(t, err)

		require.Equal(t, models.NotificationType("party_invite"), notif.Type)
		require.Equal(t, user1, notif.SenderID)
		require.Equal(t, user2, notif.RecipientID)

		var data map[string]string
		require.NoError(t, json.Unmarshal([]byte(notif.Data), &data))
		require.Equal(t, partyID, data["party_id"])
	case <-ctx.Done():
		t.Fatal("did not receive Redis notification")
	}

}
