package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	"github.com/vultr/govultr/v3"
	"golang.org/x/oauth2"
)

var (
	testService Service
	testPool    *pgxpool.Pool
	vultrClient *govultr.Client
)

func TestMain(m *testing.M) {
	envPath := filepath.Join("..", "..", "..", ".env")
	_ = godotenv.Load(envPath)

	if os.Getenv("RUN_VULTR_TESTS") != "true" {
		fmt.Println("Skipping Vultr integration tests (set RUN_VULTR_TESTS=true)")
		os.Exit(0)
	}

	apiKey := os.Getenv("VultrAPIKey")
	if apiKey == "" {
		panic("VultrAPIKey missing")
	}

	ctx := context.Background()

	var err error

	testPool, err = pgxpool.New(ctx, os.Getenv("TEST_POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	defer testPool.Close()

	config := &oauth2.Config{}
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiKey})
	vultrClient = govultr.NewClient(oauth2.NewClient(ctx, ts))

	repo := orchestratorrepo.NewOrchestratorRepository(testPool)

	testService = NewOrchestrator(repo, vultrClient)

	code := m.Run()
	os.Exit(code)
}

func TestCreateServer_init(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	instanceID, err := testService.CreateServer(ctx, "atl")
	if err != nil {
		t.Fatalf("CreateInstance failed: %v", err)
	}

	if instanceID == "" {
		t.Fatal("expected instanceID, got empty string")
	}

}
