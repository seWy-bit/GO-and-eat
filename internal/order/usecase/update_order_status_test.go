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

func TestUpdateOrderStatusUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		orderID       string
		newStatus     domain.OrderStatus
		setupMock     func(*mocks.MockOrderGetter, *mocks.MockOrderStatusUpdater)
		expectedError string
	}{
		{
			name:      "допустимый переход created -> confirmed",
			orderID:   "order-123",
			newStatus: domain.OrderStatusConfirmed,
			setupMock: func(getter *mocks.MockOrderGetter, updater *mocks.MockOrderStatusUpdater) {
				order := testOrder()
				order.Status = domain.OrderStatusCreated
				getter.EXPECT().GetOrder(mock.Anything, "order-123").Return(order, nil)
				updater.EXPECT().UpdateOrderStatus(mock.Anything, "order-123", domain.OrderStatusConfirmed).Return(nil)
			},
			expectedError: "",
		},
		{
			name:      "допустимый переход delivered -> completed",
			orderID:   "order-123",
			newStatus: domain.OrderStatusCompleted,
			setupMock: func(getter *mocks.MockOrderGetter, updater *mocks.MockOrderStatusUpdater) {
				order := testOrder()
				order.Status = domain.OrderStatusDelivered
				getter.EXPECT().GetOrder(mock.Anything, "order-123").Return(order, nil)
				updater.EXPECT().UpdateOrderStatus(mock.Anything, "order-123", domain.OrderStatusCompleted).Return(nil)
			},
			expectedError: "",
		},
		{
			name:      "недопустимый переход created -> completed",
			orderID:   "order-123",
			newStatus: domain.OrderStatusCompleted,
			setupMock: func(getter *mocks.MockOrderGetter, updater *mocks.MockOrderStatusUpdater) {
				order := testOrder()
				order.Status = domain.OrderStatusCreated
				getter.EXPECT().GetOrder(mock.Anything, "order-123").Return(order, nil)
				// UpdateOrderStatus НЕ ДОЛЖЕН вызываться!
			},
			expectedError: "invalid status transition",
		},
		{
			name:      "заказ не найден",
			orderID:   "order-not-exist",
			newStatus: domain.OrderStatusConfirmed,
			setupMock: func(getter *mocks.MockOrderGetter, updater *mocks.MockOrderStatusUpdater) {
				getter.EXPECT().GetOrder(mock.Anything, "order-not-exist").
					Return(domain.Order{}, errors.New("order not found"))
			},
			expectedError: "order not found",
		},
		{
			name:          "пустой ID",
			orderID:       "",
			newStatus:     domain.OrderStatusConfirmed,
			setupMock:     func(getter *mocks.MockOrderGetter, updater *mocks.MockOrderStatusUpdater) {},
			expectedError: "order id is required",
		},
		{
			name:          "пустой статус",
			orderID:       "order-123",
			newStatus:     "",
			setupMock:     func(getter *mocks.MockOrderGetter, updater *mocks.MockOrderStatusUpdater) {},
			expectedError: "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := mocks.NewMockOrderGetter(t)
			mockUpdater := mocks.NewMockOrderStatusUpdater(t)
			tt.setupMock(mockGetter, mockUpdater)

			uc := NewUpdateOrderStatusUseCase(mockGetter, mockUpdater)
			err := uc.Execute(context.Background(), tt.orderID, tt.newStatus)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockGetter.AssertExpectations(t)
			mockUpdater.AssertExpectations(t)
		})
	}
}
