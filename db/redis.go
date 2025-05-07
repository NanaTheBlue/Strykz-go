package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type store struct {
	client *redis.Client
}

type Store interface {
	Add(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Subscribe(ctx context.Context, channel string, handler func(message string)) error
	Publish(ctx context.Context, channel string, message []byte) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

func NewRedisInstance(redis *redis.Client) Store {
	return &store{redis}
}

// adding stuff type shit

func (s *store) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := s.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		panic(err)
	}
	return nil
}

func (s *store) Delete(ctx context.Context, key string) error {
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		panic(err)
	}
	return nil
}

func (s *store) Add(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	// prob should pass a variable for time aswell
	err := s.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		panic(err)

	}
	fmt.Println("We Wrote to that MF")
	return nil
}

func (s *store) Get(ctx context.Context, key string) (string, error) {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		panic(err)

	}
	fmt.Println("foo", val)
	return val, nil
}

// pub sub type shit
func (s *store) Publish(ctx context.Context, channel string, message []byte) error {

	err := s.client.Publish(ctx, channel, message).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *store) Subscribe(ctx context.Context, channel string, handler func(message string)) error {
	pubsub := s.client.Subscribe(ctx, channel)

	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		fmt.Println(msg.Channel, msg.Payload)
		handler(msg.Payload)
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	return nil
}

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})
	return client
}
