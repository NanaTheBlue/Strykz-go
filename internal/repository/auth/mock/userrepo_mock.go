package mock

import (
	"context"

	"github.com/nanagoboiler/models"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Check(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GrabUser(ctx context.Context, req *models.LoginRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByRefresh(ctx context.Context, refreshToken string) (*models.User, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*models.User), args.Error(1)
}
