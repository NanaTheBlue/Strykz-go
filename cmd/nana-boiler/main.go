package main

import (
	"net/http"
	"os"

	authapi "github.com/nanagoboiler/internal/api/auth"
	notificationsapi "github.com/nanagoboiler/internal/api/notifications"
	matchmakingapi "github.com/nanagoboiler/internal/api/que"
	grpcserver "github.com/nanagoboiler/internal/grpc"
	"golang.org/x/oauth2"

	"github.com/nanagoboiler/internal/bootstrap"
	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/internal/services/matchmaking"
	"github.com/nanagoboiler/internal/services/orchestrator"
	"github.com/vultr/govultr/v3"

	authrepo "github.com/nanagoboiler/internal/repository/auth"
	matchmakingrepo "github.com/nanagoboiler/internal/repository/matchmaking"
	notificationrepo "github.com/nanagoboiler/internal/repository/notification"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	redis "github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/internal/services/notifications"

	"context"
)

func main() {
	router := http.NewServeMux()
	ctx := context.Background()
	postgresURL := os.Getenv("POSTGRES_URL")
	address := os.Getenv("REDIS_ADDRESS")
	password := os.Getenv("REDIS_PASSWORD")
	apiKey := os.Getenv("VultrAPIKey")

	config := &oauth2.Config{}
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiKey})
	vultrClient := govultr.NewClient(oauth2.NewClient(ctx, ts))

	pool, err := bootstrap.NewPostgresPool(ctx, postgresURL)
	if err != nil {
		panic(err)
	}
	redisClient, err := bootstrap.NewRedisInstance(ctx, address, password)
	if err != nil {
		panic(err)
	}

	// Repositories
	authRepo := authrepo.NewUserRepository(pool)
	tokenRepo := authrepo.NewTokensRepository(pool)
	redisRepo := redis.NewRedisInstance(redisClient)
	notificationRepo := notificationrepo.NewNotificationsRepository(pool)
	orchestratorrepo := orchestratorrepo.NewOrchestratorRepository(pool)
	matchmakingRepo := matchmakingrepo.NewMatchmakingRepository(pool)

	//Connection Manager
	hub := notifications.NewHub()

	// Services
	authService := auth.NewAuthService(authRepo, tokenRepo)
	matchmakingService := matchmaking.NewMatchmakingService(redisRepo, pool, orchestratorrepo, matchmakingRepo)
	notificationService := notifications.NewnotificationsService(hub, redisRepo, notificationRepo)

	orchestrator := orchestrator.NewOrchestrator(orchestratorrepo, vultrClient)

	//grpc
	grpcserver.StartGRPC(orchestrator)

	//logger
	//logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	//middleware
	//LoggingMiddleware := middleware.LoggingMiddleware(logger)

	// Auth Handlers
	authRegister := authapi.Register(authService)
	authLogin := authapi.Login(authService)
	renew := authapi.Renew(authService)

	// Notification Handlers
	notifications := notificationsapi.Notifications(notificationService)

	//Health Handler
	health := authapi.Health()

	//MatchMaking Handlers
	inQue := matchmakingapi.Que(matchmakingService)

	//Auth Routes
	router.HandleFunc("POST /register/", authRegister)
	router.HandleFunc("POST /login/", authLogin)
	router.HandleFunc("GET /renew/", renew)

	//Health Routes
	router.HandleFunc("POST /health/", health)

	// Notification Routes
	router.HandleFunc("GET /notification/", notifications)

	//Matchmaking Routes
	router.HandleFunc("POST /que/", inQue)

	println("Server Listening on Port 8085")
	http.ListenAndServe(":8085", router)
}
