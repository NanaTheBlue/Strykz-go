package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	authrepo "github.com/nanagoboiler/internal/repository/auth"
)

var (
	testPool *pgxpool.Pool
)

func TestMain(m *testing.M) {

	fmt.Println("TEST_POSTGRES_URL:", os.Getenv("TEST_POSTGRES_URL"))

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

	code := m.Run()

	testPool.Close()

	os.Exit(code)
}

func setupService_int(pool *pgxpool.Pool) *authService {

	UserRepo := authrepo.NewUserRepository(pool)
	TokenRepo := authrepo.NewTokensRepository(pool)

	service := &authService{
		UserRepo:  UserRepo,
		TokenRepo: TokenRepo,
	}

	return service
}
