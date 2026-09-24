//go:build unit

package matchmakingapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMatchmakingService struct {
	mock.Mock
}

func (m *MockMatchmakingService) InQue(ctx context.Context, player *models.Player) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}

func (m *MockMatchmakingService) QueReader(ctx context.Context, mode string) {
	m.Called(ctx, mode)
}

func (m *MockMatchmakingService) StartMatchMaking(ctx context.Context, mode string) {
	m.Called(ctx, mode)
}

func (m *MockMatchmakingService) CreateMatch(ctx context.Context, matchCandidates []*models.Player, region string) error {
	args := m.Called(ctx, matchCandidates, region)
	return args.Error(0)
}

func (m *MockMatchmakingService) GetPlayerByID(ctx context.Context, userID string) (models.Player, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(models.Player), args.Error(1)
}

func TestQueHandler_Success(t *testing.T) {
	mockSvc := new(MockMatchmakingService)
	handler := Que(mockSvc)

	user := &models.User{ID: "user-123", Username: "testuser"}
	player := models.Player{Player_id: "user-123", Player_steamid: "steam-123"}

	mockSvc.On("GetPlayerByID", mock.Anything, "user-123").Return(player, nil)
	mockSvc.On("InQue", mock.Anything, mock.MatchedBy(func(p *models.Player) bool {
		return p.Player_id == "user-123" && p.Player_steamid == "steam-123"
	})).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/que/", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestQueHandler_Unauthorized(t *testing.T) {
	mockSvc := new(MockMatchmakingService)
	handler := Que(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/que/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockSvc.AssertNotCalled(t, "GetPlayerByID")
}

func TestQueHandler_GetPlayerError(t *testing.T) {
	mockSvc := new(MockMatchmakingService)
	handler := Que(mockSvc)

	user := &models.User{ID: "user-123"}
	mockSvc.On("GetPlayerByID", mock.Anything, "user-123").
		Return(models.Player{}, errors.New("not found"))

	req := httptest.NewRequest(http.MethodPost, "/que/", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestQueHandler_InQueError(t *testing.T) {
	mockSvc := new(MockMatchmakingService)
	handler := Que(mockSvc)

	user := &models.User{ID: "user-123"}
	player := models.Player{Player_id: "user-123", Player_steamid: "steam-123"}

	mockSvc.On("GetPlayerByID", mock.Anything, "user-123").Return(player, nil)
	mockSvc.On("InQue", mock.Anything, mock.Anything).Return(errors.New("queue full"))

	req := httptest.NewRequest(http.MethodPost, "/que/", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
