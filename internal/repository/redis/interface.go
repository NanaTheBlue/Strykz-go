package redis

import (
	"context"
	"time"

	"github.com/nanagoboiler/models"
)

type Store interface {
	Add(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Subscribe(ctx context.Context, channel string, handler func(message string)) error
	Publish(ctx context.Context, channel string, message []byte) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Count(ctx context.Context, key string) (int64, error)

	// Should maybe seperate que logic if it grows to much
	Que(ctx context.Context, mode string, region string, player *models.Player) error
	DeQue(ctx context.Context, mode, region string, count int) ([]*models.Player, error)
	DeQuePlayer(ctx context.Context, mode string, region string, playerID string) error
}
