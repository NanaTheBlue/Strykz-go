package matchmaking

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nanagoboiler/internal/bootstrap"
	matchmakingrepo "github.com/nanagoboiler/internal/repository/matchmaking"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	"github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/require"
)

var (
	testService Service
	testPool    *pgxpool.Pool
	testRedis   redis.Store
)

type fakeCapacityRequester struct{}

func (f *fakeCapacityRequester) Request(region string) {}

type fakeNotifier struct{}

func (f *fakeNotifier) CreateNoPublishNotification(ctx context.Context, notification models.Notification) (string, error) {
	return "test-id", nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	envPath := filepath.Join("..", "..", "..", ".env")
	var err error
	err = godotenv.Load(envPath)
	if err != nil {
		panic("Error loading .env file")
	}

	redisAddr := os.Getenv("TEST_REDIS_ADDR")
	redisPass := os.Getenv("TEST_REDIS_PASSWORD")

	testPool, err = pgxpool.New(ctx, os.Getenv("TEST_POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	redisClient, err := bootstrap.NewRedisInstance(ctx, redisAddr, redisPass)
	if err != nil {
		panic(err)
	}

	testRedis = redis.NewRedisInstance(redisClient)

	mmRepo := matchmakingrepo.NewMatchmakingRepository(testPool)
	orchRepo := orchestratorrepo.NewOrchestratorRepository(testPool)

	fakeCapacity := &fakeCapacityRequester{}
	fakeNotif := &fakeNotifier{}

	testService = NewMatchmakingService(
		testRedis,
		testPool,
		mmRepo,
		orchRepo,
		fakeCapacity,
		fakeNotif,
	)

	code := m.Run()

	testPool.Close()
	os.Exit(code)
}
func createTestUser(t *testing.T, pool *pgxpool.Pool, username string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE username=$1`, username)

	steamID := uuid.NewString()

	var userID string
	err := pool.QueryRow(ctx,
		`INSERT INTO users (id,username, email,hashed_password,steam_id)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id`,
		steamID,
		username,
		username+"@test.com", "TEST",
		steamID,
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

func clearTestQueue(ctx context.Context) {
	testRedis.Delete(ctx, "queue:1v1:us")

}

func TestMatchmakingCreateMatchFromQueue_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	clearTestQueue(ctx)
	_, _ = testPool.Exec(ctx, `DELETE FROM match_players`)
	_, _ = testPool.Exec(ctx, `DELETE FROM matches`)

	user1 := createTestUser(t, testPool, "mm_user1")
	user2 := createTestUser(t, testPool, "mm_user2")

	now := time.Now().Unix()
	_ = testService.InQue(ctx, &models.Player{Player_id: user1, Player_steamid: user1, JoinedAt: now, Status: "queued"})
	_ = testService.InQue(ctx, &models.Player{Player_id: user2, Player_steamid: user2, JoinedAt: now, Status: "queued"})

	go testService.StartMatchMaking(ctx, "1v1")

	var matchID string
	timeout := time.After(10 * time.Second)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			var count int
			testPool.QueryRow(ctx, "SELECT COUNT(*) FROM matches").Scan(&count)
			t.Fatalf("match was not created in time. Total matches in DB: %d", count)
		case <-tick:
			err := testPool.QueryRow(ctx,
				`SELECT match_id FROM match_players WHERE steam_id = $1 LIMIT 1`,
				user1).Scan(&matchID)

			if err == nil && matchID != "" {
				goto FOUND
			}
		}
	}
FOUND:

	var count int
	err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM match_players WHERE match_id = $1`, matchID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM match_players WHERE match_id=$1`, matchID)
		_, _ = testPool.Exec(ctx, `DELETE FROM matches WHERE id=$1`, matchID)
	})
}

func TestMatchmakingFinalizeMatch(t *testing.T) {

}
