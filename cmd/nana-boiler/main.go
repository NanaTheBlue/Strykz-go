package main

import (
	"net/http"

	authapi "github.com/nanagoboiler/internal/api/auth"
	matchmakingapi "github.com/nanagoboiler/internal/api/que"
	"github.com/nanagoboiler/internal/services/matchmaking"

	"github.com/nanagoboiler/internal/bootstrap"
	"github.com/nanagoboiler/internal/services/auth"

	authrepo "github.com/nanagoboiler/internal/repository/auth"
	redis "github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/internal/services/notifications"

	"context"
)

func main() {
	router := http.NewServeMux()
	ctx := context.Background()
	pool, err := bootstrap.NewPostgresPool(ctx)
	if err != nil {
		panic(err)
	}
	redisClient := redis.InitRedis()

	// Repositories
	authRepo := authrepo.NewUserRepository(pool)
	tokenRepo := authrepo.NewTokensRepository(pool)
	redisRepo := redis.NewRedisInstance(redisClient)

	// Services
	authService := auth.NewAuthService(authRepo, tokenRepo)
	matchmakingService := matchmaking.NewMatchmakingService(redisRepo)
	hub := notifications.NewHub()

	// Auth Handlers
	authRegister := authapi.Register(authService)
	authLogin := authapi.Login(authService)
	renew := authapi.Renew(authService)

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

	//Matchmaking Routes
	router.HandleFunc("POST /que/", inQue)

	println("Server Listening on Port 8085")
	http.ListenAndServe(":8085", router)
}
