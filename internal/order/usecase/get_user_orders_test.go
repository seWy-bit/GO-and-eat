package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/seWy-bit/GO-and-eat/internal/mocks"
	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserOrdersUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*mocks.MockUserOrdersGetter)
		expectedError  string
		expectedLength int
	}{
		{
			name:   "у пользователя есть заказы",
			userID: "user-123",
			setupMock: func(m *mocks.MockUserOrdersGetter) {
				orders := []domain.Order{testOrder()}
				m.EXPECT().GetOrdersByUser(mock.Anything, "user-123").Return(orders, nil)
			},
			expectedError:  "",
			expectedLength: 1,
		},
		{
			name:   "у пользователя нет заказов",
			userID: "user-no-orders",
			setupMock: func(m *mocks.MockUserOrdersGetter) {
				m.EXPECT().GetOrdersByUser(mock.Anything, "user-no-orders").Return([]domain.Order{}, nil)
			},
			expectedError:  "",
			expectedLength: 0,
		},
		{
			name:   "ошибка базы данных",
			userID: "user-123",
			setupMock: func(m *mocks.MockUserOrdersGetter) {
				m.EXPECT().GetOrdersByUser(mock.Anything, "user-123").Return([]domain.Order{}, errors.New("database error"))
			},
			expectedError:  "database error",
			expectedLength: 0,
		},
		{
			name:           "пустой ID пользователя",
			userID:         "",
			setupMock:      func(m *mocks.MockUserOrdersGetter) {},
			expectedError:  "user id is required",
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := mocks.NewMockUserOrdersGetter(t)
			tt.setupMock(mockGetter)

			uc := NewGetUserOrdersUseCase(mockGetter)
			orders, err := uc.Execute(context.Background(), tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, orders, tt.expectedLength)
			}

			mockGetter.AssertExpectations(t)
		})
	}
}
