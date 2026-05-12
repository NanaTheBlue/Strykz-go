//go:build integration

package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	"github.com/vultr/govultr/v3"
)

var (
	testService Service
	testPool    *pgxpool.Pool
	vultrClient *govultr.Client
)

func TestMain(m *testing.M) {
	envPath := filepath.Join("..", "..", "..", ".env")
	_ = godotenv.Load(envPath)

	if os.Getenv("RUN_VULTRls_TESTS") != "true" {
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

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	repo := orchestratorrepo.NewOrchestratorRepository(testPool)

	testService = NewOrchestrator(repo, ec2Client)

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
