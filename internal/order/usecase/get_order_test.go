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

func TestGetOrderUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		orderID       string
		setupMock     func(*mocks.MockOrderGetter)
		expectedError string
		validateOrder func(*testing.T, *domain.Order)
	}{
		{
			name:    "успешное получение заказа",
			orderID: "order-123",
			setupMock: func(m *mocks.MockOrderGetter) {
				m.EXPECT().GetOrder(mock.Anything, "order-123").Return(testOrder(), nil)
			},
			expectedError: "",
			validateOrder: func(t *testing.T, order *domain.Order) {
				assert.Equal(t, "order-123", order.ID)
				assert.Equal(t, "user-123", order.UserID)
			},
		},
		{
			name:    "заказ не найден",
			orderID: "order-not-exist",
			setupMock: func(m *mocks.MockOrderGetter) {
				m.EXPECT().
					GetOrder(mock.Anything, "order-not-exist").
					Return(domain.Order{}, errors.New("order not found"))
			},
			expectedError: "order not found",
			validateOrder: nil,
		},
		{
			name:    "ошибка базы данных",
			orderID: "order-123",
			setupMock: func(m *mocks.MockOrderGetter) {
				m.EXPECT().
					GetOrder(mock.Anything, "order-123").
					Return(domain.Order{}, errors.New("database connection failed"))
			},
			expectedError: "database connection failed",
			validateOrder: nil,
		},
		{
			name:          "пустой ID",
			orderID:       "",
			setupMock:     func(m *mocks.MockOrderGetter) {},
			expectedError: "order id is required",
			validateOrder: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := mocks.NewMockOrderGetter(t)
			tt.setupMock(mockGetter)

			uc := NewGetOrderUseCase(mockGetter)

			order, err := uc.Execute(context.Background(), tt.orderID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				tt.validateOrder(t, order)
			}

			mockGetter.AssertExpectations(t)
		})
	}
}
