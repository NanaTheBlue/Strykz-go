package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockTokenRepo struct {
	mock.Mock
}

func (m *MockTokenRepo) AddRefresh(ctx context.Context, jti, userID string) error {
	args := m.Called(ctx, jti, userID)
	return args.Error(0)
}
