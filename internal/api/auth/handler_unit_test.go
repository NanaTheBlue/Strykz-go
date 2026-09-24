//go:build unit

package authapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(ctx context.Context, req *models.RegisterRequest) (models.Tokens, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.Tokens), args.Error(1)
}

func (m *MockAuthService) LoginUser(ctx context.Context, req *models.LoginRequest) (models.Tokens, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.Tokens), args.Error(1)
}

func (m *MockAuthService) RenewToken(ctx context.Context, refreshToken string) (models.Tokens, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(models.Tokens), args.Error(1)
}

func TestHealthHandler(t *testing.T) {
	handler := Health()
	req := httptest.NewRequest(http.MethodPost, "/health/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRegisterHandler_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Register(mockSvc)

	body := models.RegisterRequest{
		Username:        "testuser",
		Email:           "test@test.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	bodyBytes, _ := json.Marshal(body)

	mockSvc.On("RegisterUser", mock.Anything, mock.MatchedBy(func(r *models.RegisterRequest) bool {
		return r.Username == "testuser" && r.Email == "test@test.com"
	})).Return(models.Tokens{Auth_token: "auth", Refresh_token: "refresh"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Register(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "RegisterUser")
}

func TestRegisterHandler_ValidationFails(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Register(mockSvc)

	body := models.RegisterRequest{
		Username:        "ab",
		Email:           "bad",
		Password:        "short",
		ConfirmPassword: "different",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "RegisterUser")
}

func TestRegisterHandler_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Register(mockSvc)

	body := models.RegisterRequest{
		Username:        "testuser",
		Email:           "test@test.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	bodyBytes, _ := json.Marshal(body)

	mockSvc.On("RegisterUser", mock.Anything, mock.Anything).
		Return(models.Tokens{}, errors.New("db error"))

	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Login(mockSvc)

	body := models.LoginRequest{Email: "test@test.com", Password: "password123"}
	bodyBytes, _ := json.Marshal(body)

	mockSvc.On("LoginUser", mock.Anything, mock.MatchedBy(func(r *models.LoginRequest) bool {
		return r.Email == "test@test.com"
	})).Return(models.Tokens{Auth_token: "auth", Refresh_token: "refresh"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/login/", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Login(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/login/", bytes.NewReader([]byte("{bad")))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "LoginUser")
}

func TestLoginHandler_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Login(mockSvc)

	body := models.LoginRequest{Email: "test@test.com", Password: "wrong"}
	bodyBytes, _ := json.Marshal(body)

	mockSvc.On("LoginUser", mock.Anything, mock.Anything).
		Return(models.Tokens{}, errors.New("invalid credentials"))

	req := httptest.NewRequest(http.MethodPost, "/login/", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRenewHandler_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Renew(mockSvc)

	mockSvc.On("RenewToken", mock.Anything, mock.Anything).
		Return(models.Tokens{Auth_token: "new-auth", Refresh_token: "new-refresh"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/renew/", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid-refresh"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestRenewHandler_NoCookie(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Renew(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/renew/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "RenewToken")
}

func TestRenewHandler_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := Renew(mockSvc)

	mockSvc.On("RenewToken", mock.Anything, mock.Anything).
		Return(models.Tokens{}, errors.New("token expired"))

	req := httptest.NewRequest(http.MethodGet, "/renew/", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "testuser", false},
		{"valid with numbers", "user123", false},
		{"valid with underscore", "test_user", false},
		{"valid with dash", "test-user", false},
		{"too short", "ab", true},
		{"too long", "thisusernameistoolongfortheregex", true},
		{"invalid chars", "test user!", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pass    string
		confirm string
		wantErr bool
	}{
		{"valid", "password123", "password123", false},
		{"too short", "short", "short", true},
		{"too long", "thispasswordistoolongforval", "thispasswordistoolongforval", true},
		{"mismatch", "password123", "password456", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.pass, tt.confirm)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "test@test.com", false},
		{"valid with dot", "user.name@test.com", false},
		{"valid subdomain", "user@sub.test.com", false},
		{"no at", "testtest.com", true},
		{"no domain", "test@", true},
		{"no tld", "test@test", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRegistration(t *testing.T) {
	validReq := &models.RegisterRequest{
		Username:        "testuser",
		Email:           "test@test.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	assert.NoError(t, validateRegistration(validReq))

	invalidReq := &models.RegisterRequest{
		Username:        "ab",
		Email:           "bad",
		Password:        "short",
		ConfirmPassword: "different",
	}
	assert.Error(t, validateRegistration(invalidReq))
}

var _ auth.Service = (*MockAuthService)(nil)
