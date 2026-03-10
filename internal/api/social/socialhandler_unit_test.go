//go:build unit

package socialapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockSocialService struct {
	mock.Mock
}

func (m *MockSocialService) AcceptFriendRequest(ctx context.Context, userID string, requestID string) error {
	args := m.Called(ctx, userID, requestID)
	return args.Error(0)
}

func (m *MockSocialService) BlockUser(ctx context.Context, userID, targetID string) error {
	args := m.Called(ctx, userID, targetID)
	return args.Error(0)
}

func TestAcceptFriendRequestHandler(t *testing.T) {

	validUUID := uuid.New().String()
	serviceErr := errors.New("service failed")

	tests := []struct {
		name         string
		user         *models.User
		reqID        string
		mockReturn   error
		wantStatus   int
		expectCalled bool
	}{
		{
			name:         "success",
			user:         &models.User{ID: "user-123"},
			reqID:        validUUID,
			mockReturn:   nil,
			wantStatus:   http.StatusNoContent,
			expectCalled: true,
		},
		{
			name:         "service error",
			user:         &models.User{ID: "user-123"},
			reqID:        validUUID,
			mockReturn:   serviceErr,
			wantStatus:   http.StatusInternalServerError,
			expectCalled: true,
		},
		{
			name:         "unauthorized",
			user:         nil,
			reqID:        validUUID,
			mockReturn:   nil,
			wantStatus:   http.StatusUnauthorized,
			expectCalled: false,
		},
		{
			name:         "invalid UUID",
			user:         &models.User{ID: "user-123"},
			reqID:        "bad-uuid",
			mockReturn:   nil,
			wantStatus:   http.StatusBadRequest,
			expectCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSocialService{}
			if tt.expectCalled {
				mockService.On("AcceptFriendRequest", mock.Anything, tt.user.ID, tt.reqID).Return(tt.mockReturn)
			}

			handler := AcceptFriendRequest(mockService)

			req := httptest.NewRequest(
				http.MethodPost,
				"/friend-requests/"+tt.reqID+"/accept",
				nil,
			)
			req.SetPathValue("id", tt.reqID)

			if tt.user != nil {
				ctx := context.WithValue(req.Context(), auth.UserContextKey, tt.user)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("[%s] expected status %d got %d", tt.name, tt.wantStatus, rec.Code)
			}

			if tt.expectCalled {
				mockService.AssertCalled(t, "AcceptFriendRequest", mock.Anything, tt.user.ID, tt.reqID)
			} else {
				if len(mockService.Calls) != 0 {
					t.Fatalf("[%s] service should not be called", tt.name)
				}
			}
		})
	}
}
