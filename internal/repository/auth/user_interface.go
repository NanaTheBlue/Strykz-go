package authrepo

import (
	"context"

	"github.com/nanagoboiler/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Check(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, user *models.User) error
	GrabUser(ctx context.Context, req *models.LoginRequest) (*models.User, error)
	GetUserByRefresh(ctx context.Context, refreshToken string) (*models.User, error)
}
