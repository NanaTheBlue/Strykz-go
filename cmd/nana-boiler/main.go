package main

import (
	"log"
	"net/http"
	"os"
	"time"

	authapi "github.com/nanagoboiler/internal/api/auth"
	"github.com/nanagoboiler/internal/api/middleware"
	notificationsapi "github.com/nanagoboiler/internal/api/notifications"
	matchmakingapi "github.com/nanagoboiler/internal/api/que"
	socialapi "github.com/nanagoboiler/internal/api/social"
	grpcserver "github.com/nanagoboiler/internal/grpc"
	"golang.org/x/oauth2"

	"github.com/nanagoboiler/internal/bootstrap"
	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/internal/services/matchmaking"
	"github.com/nanagoboiler/internal/services/orchestrator"
	"github.com/nanagoboiler/internal/services/social"
	"github.com/vultr/govultr/v3"

	authrepo "github.com/nanagoboiler/internal/repository/auth"
	matchmakingrepo "github.com/nanagoboiler/internal/repository/matchmaking"
	notificationrepo "github.com/nanagoboiler/internal/repository/notification"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	redis "github.com/nanagoboiler/internal/repository/redis"
	socialrepo "github.com/nanagoboiler/internal/repository/social"
	"github.com/nanagoboiler/internal/services/notifications"

	"context"
)

func main() {
	router := http.NewServeMux()
	rl := middleware.NewRateLimiter(10, time.Second*1)
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
		log.Fatalf("failed to initialize Postgres: %v", err)
	}
	redisClient, err := bootstrap.NewRedisInstance(ctx, address, password)
	if err != nil {
		log.Fatalf("failed to initialize Redis: %v", err)
	}

	// Repositories
	authRepo := authrepo.NewUserRepository(pool)
	tokenRepo := authrepo.NewTokensRepository(pool)
	redisRepo := redis.NewRedisInstance(redisClient)
	notificationRepo := notificationrepo.NewNotificationsRepository(pool)
	orchestratorrepo := orchestratorrepo.NewOrchestratorRepository(pool)
	matchmakingRepo := matchmakingrepo.NewMatchmakingRepository(pool)
	socialRepo := socialrepo.NewSocialRepository(pool)

	//Connection Manager
	hub := notifications.NewHub()

	// Services
	authService := auth.NewAuthService(authRepo, tokenRepo)
	orchestrator := orchestrator.NewOrchestrator(orchestratorrepo, vultrClient)
	notificationService := notifications.NewnotificationsService(hub, redisRepo, notificationRepo)
	matchmakingService := matchmaking.NewMatchmakingService(redisRepo, pool, matchmakingRepo, orchestratorrepo, orchestrator, notificationService, orchestrator)
	socialService := social.NewsocialService(notificationService, pool, socialRepo, redisRepo)

	//grpc
	grpcserver.StartGRPC(orchestrator, ":6767")

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

	// Social Handlers
	blockUser := socialapi.BlockUser(socialService)
	acceptFriendRequest := socialapi.AcceptFriendRequest(socialService)

	//Health Handler
	health := authapi.Health()

	//MatchMaking Handlers
	inQue := matchmakingapi.Que(matchmakingService)

	//Auth Routes
	router.HandleFunc("POST /register/", middleware.CORSMiddleware(rl.Limit(authRegister)))
	router.HandleFunc("POST /login/", middleware.CORSMiddleware(rl.Limit(authLogin)))
	router.HandleFunc("GET /renew/", middleware.CORSMiddleware(rl.Limit(renew)))

	// Social Routes
	router.HandleFunc("POST /block/", middleware.CORSMiddleware(rl.Limit(middleware.AuthMiddleware(blockUser))))
	router.HandleFunc("POST /friend-requests/{id}/accept", middleware.CORSMiddleware(rl.Limit(middleware.AuthMiddleware(acceptFriendRequest))))

	//Health Routes
	router.HandleFunc("POST /health/", middleware.CORSMiddleware(rl.Limit(health)))

	// Notification Routes
	router.HandleFunc("GET /notification/", middleware.CORSMiddleware(rl.Limit(middleware.AuthMiddleware(notifications))))

	//Matchmaking Routes
	router.HandleFunc("POST /que/", middleware.CORSMiddleware(rl.Limit(middleware.AuthMiddleware(inQue))))

	println("Server Listening on Port 8080")
	http.ListenAndServe(":8080", router)
}
