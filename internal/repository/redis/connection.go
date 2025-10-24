package redis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {

	address := os.Getenv("REDIS_ADDRESS")
	password := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0, // Use default DB
		Protocol: 2, // Connection protocol
	})
	return client
}
