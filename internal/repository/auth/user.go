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

func (r *userRepo) Check(ctx context.Context, user *models.User) error {
	var id string
	err := r.pool.QueryRow(ctx, "SELECT * from users WHERE refresh_token = $1", user.ID).Scan(&id)

	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) GetUserByRefresh(ctx context.Context, refreshToken string) (*models.User, error) {
	var user models.User

	err := r.pool.QueryRow(ctx, "SELECT id from users WHERE Id = $1", refreshToken).Scan(&user)

	if err != nil {
		return &models.User{}, err
	}

	return &user, nil

}

func (r *userRepo) GrabUser(ctx context.Context, req *models.LoginRequest) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, "SELECT id,username,email,password_hash from users WHERE email = $1", user.Email).Scan(&user)
	if err != nil {
		return &models.User{}, err
	}
	return &user, nil
}
