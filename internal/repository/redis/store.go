package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nanagoboiler/models"
	"github.com/redis/go-redis/v9"
)

type store struct {
	client *redis.Client
}

func NewRedisInstance(redis *redis.Client) Store {
	return &store{redis}
}

func (s *store) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := s.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		panic(err)
	}
	return nil
}

func (s *store) Count(ctx context.Context, key string) (int64, error) {
	count, err := s.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *store) Delete(ctx context.Context, key string) error {
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		panic(err)
	}
	return nil
}

func (s *store) AddNX(ctx context.Context, key string, value string, exp time.Duration) (bool, error) {

	cmd := s.client.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: "NX",
		TTL:  exp,
	})

	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *store) Add(ctx context.Context, key string, value []byte, expiration time.Duration) error {

	err := s.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		panic(err)

	}
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

func (s *store) Publish(ctx context.Context, channel string, message models.Notification) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return s.client.Publish(ctx, channel, payload).Err()

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

func (s *store) Que(ctx context.Context, mode string, region string, player *models.Player) error {
	playerBytes, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("failed to serialize player: %w", err)
	}

	queueKey := fmt.Sprintf("queue:%s:%s", mode, region)
	indexKey := fmt.Sprintf("queue_index:%s", player.Player_id)

	pipe := s.client.TxPipeline()

	pipe.ZAdd(ctx, queueKey, redis.Z{
		Score:  float64(player.JoinedAt),
		Member: string(playerBytes),
	})

	pipe.Set(ctx, indexKey, string(playerBytes), 30*time.Minute)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to queue player: %w", err)
	}

	return nil

}

func (s *store) DeQue(ctx context.Context, mode string, region string, count int) ([]*models.Player, error) {
	queueKey := fmt.Sprintf("queue:%s:%s", mode, region)

	playerBytes, err := s.client.ZRange(ctx, queueKey, 0, int64(count-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from queue: %w", err)
	}

	if len(playerBytes) == 0 {
		return nil, nil
	}

	if err := s.client.ZRemRangeByRank(ctx, queueKey, 0, int64(count-1)).Err(); err != nil {
		return nil, fmt.Errorf("failed to remove players from queue: %w", err)
	}

	var players []*models.Player
	for _, pb := range playerBytes {
		var p models.Player
		if err := json.Unmarshal([]byte(pb), &p); err != nil {
			continue
		}
		players = append(players, &p)
	}

	// adds the Players back in if count is less than needed
	if len(players) < count {
		pipe := s.client.Pipeline()
		for _, p := range players {
			playerJSON, _ := json.Marshal(p)
			pipe.ZAdd(ctx, queueKey, redis.Z{
				Score:  float64(time.Now().Unix()),
				Member: playerJSON,
			})
		}
		return nil, nil
	}

	return players, nil

}

func (s *store) DeQuePlayer(ctx context.Context, mode string, region string, playerID string) error {
	queueKey := fmt.Sprintf("queue:%s:%s", mode, region)
	indexKey := fmt.Sprintf("queue_index:%s", playerID)

	member, err := s.client.Get(ctx, indexKey).Result()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to lookup player: %w", err)
	}

	if err := s.client.ZRem(ctx, queueKey, member).Err(); err != nil {
		return fmt.Errorf("failed to remove player from queue: %w", err)
	}

	s.client.Del(ctx, indexKey)

	return nil
}
