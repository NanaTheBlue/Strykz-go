package authrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nanagoboiler/models"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	_, err := r.pool.Exec(ctx, "INSERT into users (email,username,hashed_password) VALUES ($1, $2, $3);", user.Email, user.Username, user.PasswordHash)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) Delete(ctx context.Context, user *models.User) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepo) GetUserByRefresh(ctx context.Context, refreshToken string) (*models.User, error) {
	var user models.User

	err := r.pool.QueryRow(ctx, "SELECT id,username,email from users WHERE refresh_token = $1", refreshToken).Scan(&user.ID, &user.Username, &user.Email)

	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (r *userRepo) GrabUser(ctx context.Context, req *models.LoginRequest) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, "SELECT id,username,email,password_hash from users WHERE email = $1", req.Email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
