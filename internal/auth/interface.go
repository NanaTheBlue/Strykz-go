package auth

import (
	"context"

	"github.com/nanagoboiler/models"
)

type Service interface {
	RegisterUser(ctx context.Context, req *models.RegisterRequest) (models.Tokens, error)
	LoginUser(ctx context.Context, req *models.LoginRequest) (models.Tokens, error)
	RenewToken(ctx context.Context, refreshToken string) (models.Tokens, error)
}
