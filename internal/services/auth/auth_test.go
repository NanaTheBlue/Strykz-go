//go:build unit

package auth

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authMocks "github.com/nanagoboiler/internal/repository/auth/mock"
	"github.com/nanagoboiler/models"
)

func setupService() *authService {
	os.Setenv("JWT_SECRET", "testsecret")

	userRepo := new(authMocks.MockUserRepo)
	tokenRepo := new(authMocks.MockTokenRepo)

	return &authService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
		secret:    "testsecret",
	}
}

func TestRegisterUser_Success(t *testing.T) {
	svc := setupService()

	req := &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@test.com",
		Password: "password123",
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
			return u.Email == req.Email &&
				u.Username == req.Username &&
				u.PasswordHash != ""
		})).
		Return(nil)

	svc.TokenRepo.(*authMocks.MockTokenRepo).
		On("AddRefresh", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	tokens, err := svc.RegisterUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.Auth_token)
	assert.NotEmpty(t, tokens.Refresh_token)

	svc.UserRepo.(*authMocks.MockUserRepo).AssertExpectations(t)
	svc.TokenRepo.(*authMocks.MockTokenRepo).AssertExpectations(t)
}

func TestLoginUser_Success(t *testing.T) {
	svc := setupService()

	password := "password123"
	hash, err := HashPassword([]byte(password))
	assert.NoError(t, err)

	req := &models.LoginRequest{
		Email:    "test@test.com",
		Password: password,
	}

	user := &models.User{
		ID:           uuid.NewString(),
		Username:     "testuser",
		Email:        req.Email,
		PasswordHash: hash,
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("GrabUser", mock.Anything, mock.MatchedBy(func(r *models.LoginRequest) bool {
			return r.Email == req.Email
		})).
		Return(user, nil)

	svc.TokenRepo.(*authMocks.MockTokenRepo).
		On("AddRefresh", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	tokens, err := svc.LoginUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.Auth_token)
	assert.NotEmpty(t, tokens.Refresh_token)

	svc.UserRepo.(*authMocks.MockUserRepo).AssertExpectations(t)
	svc.TokenRepo.(*authMocks.MockTokenRepo).AssertExpectations(t)
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	svc := setupService()

	correctPassword := "password123"
	hash, err := HashPassword([]byte(correctPassword))
	assert.NoError(t, err)

	req := &models.LoginRequest{
		Email:    "test@test.com",
		Password: "wrongpassword",
	}

	user := &models.User{
		ID:           uuid.NewString(),
		Username:     "testuser",
		Email:        req.Email,
		PasswordHash: hash,
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("GrabUser", mock.Anything, mock.Anything).
		Return(user, nil)

	_, err = svc.LoginUser(context.Background(), req)
	assert.Error(t, err)

	svc.UserRepo.(*authMocks.MockUserRepo).AssertExpectations(t)
}

func TestRenewToken_UserNotFound(t *testing.T) {
	svc := setupService()

	refreshToken := "fake-refresh-token"

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("GetUserByRefresh", mock.Anything, refreshToken).
		Return(&models.User{}, errors.New("not found"))

	_, err := svc.RenewToken(context.Background(), refreshToken)
	assert.Error(t, err)

	svc.UserRepo.(*authMocks.MockUserRepo).AssertExpectations(t)
}

func TestGenerateTokens_ValidJWT(t *testing.T) {
	svc := setupService()

	user := &models.User{
		ID:       uuid.NewString(),
		Username: "testuser",
	}

	tokens, jti, err := svc.generateTokens(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.Auth_token)
	assert.NotEmpty(t, tokens.Refresh_token)
	assert.NotEmpty(t, jti)

	parsedToken, err := jwt.Parse(tokens.Auth_token, func(t *jwt.Token) (interface{}, error) {
		return []byte(svc.secret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

func TestRegisterUser_UserCreationError(t *testing.T) {
	svc := setupService()

	req := &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@test.com",
		Password: "password123",
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("Create", mock.Anything, mock.Anything).
		Return(errors.New("user already exists"))

	_, err := svc.RegisterUser(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
}

func TestRegisterUser_TokenAdditionError(t *testing.T) {
	svc := setupService()

	req := &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@test.com",
		Password: "password123",
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("Create", mock.Anything, mock.Anything).
		Return(nil)

	svc.TokenRepo.(*authMocks.MockTokenRepo).
		On("AddRefresh", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("token storage failed"))

	_, err := svc.RegisterUser(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "token storage failed", err.Error())
}

func TestLoginUser_UserNotFound(t *testing.T) {
	svc := setupService()

	req := &models.LoginRequest{
		Email:    "nonexistent@test.com",
		Password: "password123",
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("GrabUser", mock.Anything, mock.Anything).
		Return(nil, errors.New("user not found"))

	_, err := svc.LoginUser(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestLoginUser_TokenAdditionError(t *testing.T) {
	svc := setupService()

	password := "password123"
	hash, err := HashPassword([]byte(password))
	assert.NoError(t, err)

	req := &models.LoginRequest{
		Email:    "test@test.com",
		Password: password,
	}

	user := &models.User{
		ID:           uuid.NewString(),
		Username:     "testuser",
		Email:        req.Email,
		PasswordHash: hash,
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("GrabUser", mock.Anything, mock.Anything).
		Return(user, nil)

	svc.TokenRepo.(*authMocks.MockTokenRepo).
		On("AddRefresh", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("token storage error"))

	_, err = svc.LoginUser(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "token storage error", err.Error())
}

func TestRenewToken_Success(t *testing.T) {
	svc := setupService()

	refreshToken := "valid-refresh-token"
	user := &models.User{
		ID:       uuid.NewString(),
		Username: "testuser",
		Email:    "test@test.com",
	}

	svc.UserRepo.(*authMocks.MockUserRepo).
		On("GetUserByRefresh", mock.Anything, refreshToken).
		Return(user, nil)

	tokens, err := svc.RenewToken(context.Background(), refreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.Auth_token)
	assert.NotEmpty(t, tokens.Refresh_token)

	svc.UserRepo.(*authMocks.MockUserRepo).AssertExpectations(t)
}

func TestGenerateTokens_ContainsValidClaims(t *testing.T) {
	svc := setupService()

	user := &models.User{
		ID:       uuid.NewString(),
		Username: "claimsuser",
		Email:    "claims@test.com",
	}

	tokens, jti, err := svc.generateTokens(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, jti)

	// Parse and verify auth token claims
	parsedToken, err := jwt.Parse(tokens.Auth_token, func(t *jwt.Token) (interface{}, error) {
		return []byte(svc.secret), nil
	})

	assert.NoError(t, err)
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, user.Username, claims["userName"])
	assert.Equal(t, user.ID, claims["userId"])
	assert.NotNil(t, claims["exp"])
	assert.NotNil(t, claims["iat"])

	// Parse and verify refresh token claims
	parsedRefresh, err := jwt.Parse(tokens.Refresh_token, func(t *jwt.Token) (interface{}, error) {
		return []byte(svc.secret), nil
	})

	assert.NoError(t, err)
	refreshClaims, ok := parsedRefresh.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, user.Username, refreshClaims["userName"])
	assert.Equal(t, user.ID, refreshClaims["userId"])
	assert.Equal(t, jti, refreshClaims["jti"])
	assert.NotNil(t, refreshClaims["exp"])
	assert.NotNil(t, refreshClaims["iat"])
}

func TestHashPassword_Success(t *testing.T) {
	password := []byte("password123")

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, string(password), hash)
}

func TestValidateHashedPassword_Success(t *testing.T) {
	password := "password123"

	hash, err := HashPassword([]byte(password))
	assert.NoError(t, err)

	err = validateHashedPassword(password, hash)

	assert.NoError(t, err)
}

func TestValidateHashedPassword_InvalidPassword(t *testing.T) {
	password := "password123"
	wrongPassword := "wrongpassword"

	hash, err := HashPassword([]byte(password))
	assert.NoError(t, err)

	err = validateHashedPassword(wrongPassword, hash)

	assert.Error(t, err)
}
