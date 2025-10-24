package auth

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	authrepo "github.com/nanagoboiler/internal/repository/auth"
	"github.com/nanagoboiler/models"
)

type authService struct {
	UserRepo  authrepo.UserRepository
	TokenRepo authrepo.TokensRepository
	secret    string
}

func NewAuthService(userrepo authrepo.UserRepository, tokensrepo authrepo.TokensRepository) Service {
	return &authService{UserRepo: userrepo, TokenRepo: tokensrepo, secret: os.Getenv("JWT_SECRET")}
}

func (s *authService) RenewToken(ctx context.Context, refreshToken string) (models.Tokens, error) {

	userById, err := s.UserRepo.GetUserByRefresh(ctx, refreshToken)
	if err != nil {
		return models.Tokens{}, err
	}

	user := models.User{
		ID:           userById.ID,
		Username:     userById.Username,
		Email:        userById.Email,
		PasswordHash: "",
	}

	token, _, err := s.generateTokens(&user)
	if err != nil {
		return models.Tokens{}, err
	}

	return token, nil
}

func (s *authService) RegisterUser(ctx context.Context, req *models.RegisterRequest) (models.Tokens, error) {

	passwordHash, err := HashPassword([]byte(req.Password))
	if err != nil {
		return models.Tokens{}, err
	}

	user := models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	err = s.UserRepo.Create(ctx, &user)
	if err != nil {
		return models.Tokens{}, err
	}
	token, jti, err := s.generateTokens(&user)
	if err != nil {
		return models.Tokens{}, err
	}
	err = s.TokenRepo.AddRefresh(ctx, jti, user.ID)
	if err != nil {
		return models.Tokens{}, err
	}

	return token, nil
}

func (s *authService) LoginUser(ctx context.Context, req *models.LoginRequest) (models.Tokens, error) {
	user, err := s.UserRepo.GrabUser(ctx, req)
	if err != nil {
		return models.Tokens{}, err
	}
	err = validateHashedPassword(req.Password, user.PasswordHash)
	if err != nil {
		return models.Tokens{}, err
	}
	tokens, jti, err := s.generateTokens(user)
	if err != nil {
		return models.Tokens{}, err
	}
	err = s.TokenRepo.AddRefresh(ctx, jti, user.ID)
	if err != nil {
		return models.Tokens{}, err
	}

	return tokens, nil
}

func (s *authService) generateTokens(user *models.User) (token models.Tokens, jti string, err error) {
	jti = uuid.NewString()
	now := time.Now()
	auth_claims := jwt.MapClaims{
		"userName": user.Username,
		"userId":   user.ID,
		"exp":      now.Add(10 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	}

	refresh_claims := jwt.MapClaims{
		"userName": user.Username,
		"userId":   user.ID,
		"exp":      now.Add(30 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"jti":      jti,
	}

	authTok := jwt.NewWithClaims(jwt.SigningMethodHS256, auth_claims)

	authToken, err := authTok.SignedString([]byte(s.secret))
	if err != nil {
		return token, "", err
	}
	refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refresh_claims)

	refreshToken, err := refreshTok.SignedString([]byte(s.secret))
	if err != nil {
		return token, "", err
	}

	token = models.Tokens{
		Auth_token:    authToken,
		Refresh_token: refreshToken,
	}

	return token, jti, nil

}
