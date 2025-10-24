package authrepo

import (
	"context"
)

type TokensRepository interface {
	AddRefresh(ctx context.Context, jti string, uuid string) error
}
