package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/seWy-bit/GO-and-eat/internal/mocks"
	restaurantDomain "github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrderUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateOrderInput
		setupMock func(
			*mocks.MockOrderCreator,
			*mocks.MockMenuGetter,
			*mocks.MockStockChecker,
			*mocks.MockStockDecreaser,
			*mocks.MockTransactionManager,
		)
		expectedError string
	}{
		{
			name:  "успешное создание заказа",
			input: testCreateOrderInput(),
			setupMock: func(
				creator *mocks.MockOrderCreator,
				menuGetter *mocks.MockMenuGetter,
				stockChecker *mocks.MockStockChecker,
				stockDecreaser *mocks.MockStockDecreaser,
				txManager *mocks.MockTransactionManager,
			) {
				// Мок для транзакции
				var mockTx pgx.Tx
				txManager.EXPECT().Begin(mock.Anything).Return(mockTx, nil)
				txManager.EXPECT().Commit(mockTx).Return(nil)
				txManager.EXPECT().Rollback(mockTx).Return(nil)

				// Мок для получения меню
				menuGetter.EXPECT().GetMenu("rest-123").Return(testMenu(), nil)

				// Мок для проверки наличия
				stockChecker.EXPECT().CheckAvailabilityWithTx(
					mock.Anything, mockTx, "rest-123", mock.Anything,
				).Return(true, nil)

				// Мок для уменьшения stock
				stockDecreaser.EXPECT().DecreaseStockWithTx(
					mock.Anything, mockTx, "rest-123", "pizza-1", 2,
				).Return(nil)

				// Мок для сохранения заказа
				creator.EXPECT().CreateOrderWithTx(mock.Anything, mockTx, mock.Anything).Return(nil)
			},
			expectedError: "",
		},
		{
			name:  "недостаточно stock",
			input: testCreateOrderInput(),
			setupMock: func(
				creator *mocks.MockOrderCreator,
				menuGetter *mocks.MockMenuGetter,
				stockChecker *mocks.MockStockChecker,
				stockDecreaser *mocks.MockStockDecreaser,
				txManager *mocks.MockTransactionManager,
			) {
				var mockTx pgx.Tx
				txManager.EXPECT().Begin(mock.Anything).Return(mockTx, nil)
				txManager.EXPECT().Rollback(mockTx).Return(nil)

				menuGetter.EXPECT().GetMenu("rest-123").Return(testMenu(), nil)
				stockChecker.EXPECT().CheckAvailabilityWithTx(
					mock.Anything, mockTx, "rest-123", mock.Anything,
				).Return(false, nil)

				// DecreaseStockWithTx и CreateOrderWithTx НЕ должны вызываться!
			},
			expectedError: "not enough stock",
		},
		{
			name:  "блюдо не найдено в меню",
			input: testCreateOrderInput(),
			setupMock: func(
				creator *mocks.MockOrderCreator,
				menuGetter *mocks.MockMenuGetter,
				stockChecker *mocks.MockStockChecker,
				stockDecreaser *mocks.MockStockDecreaser,
				txManager *mocks.MockTransactionManager,
			) {
				var mockTx pgx.Tx
				txManager.EXPECT().Begin(mock.Anything).Return(mockTx, nil)
				txManager.EXPECT().Rollback(mockTx).Return(nil)

				// Меню не содержит pizza-1
				emptyMenu := []restaurantDomain.MenuItem{}
				menuGetter.EXPECT().GetMenu("rest-123").Return(emptyMenu, nil)
			},
			expectedError: "menu item not found",
		},
		{
			name:  "ошибка при сохранении заказа",
			input: testCreateOrderInput(),
			setupMock: func(
				creator *mocks.MockOrderCreator,
				menuGetter *mocks.MockMenuGetter,
				stockChecker *mocks.MockStockChecker,
				stockDecreaser *mocks.MockStockDecreaser,
				txManager *mocks.MockTransactionManager,
			) {
				var mockTx pgx.Tx
				txManager.EXPECT().Begin(mock.Anything).Return(mockTx, nil)
				txManager.EXPECT().Rollback(mockTx).Return(nil)

				menuGetter.EXPECT().GetMenu("rest-123").Return(testMenu(), nil)
				stockChecker.EXPECT().CheckAvailabilityWithTx(
					mock.Anything, mockTx, "rest-123", mock.Anything,
				).Return(true, nil)
				stockDecreaser.EXPECT().DecreaseStockWithTx(
					mock.Anything, mockTx, "rest-123", "pizza-1", 2,
				).Return(nil)
				creator.EXPECT().CreateOrderWithTx(mock.Anything, mockTx, mock.Anything).
					Return(errors.New("database error"))
			},
			expectedError: "failed to save order",
		},
		{
			name: "пустой ID",
			input: CreateOrderInput{
				ID:           "",
				UserID:       "user-123",
				RestaurantID: "rest-123",
				Items:        []OrderItemInput{{MenuItemID: "pizza-1", Quantity: 1}},
			},
			setupMock: func(creator *mocks.MockOrderCreator, menuGetter *mocks.MockMenuGetter, stockChecker *mocks.MockStockChecker, stockDecreaser *mocks.MockStockDecreaser, txManager *mocks.MockTransactionManager) {
			},
			expectedError: "id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём моки для всех зависимостей
			mockCreator := mocks.NewMockOrderCreator(t)
			mockMenuGetter := mocks.NewMockMenuGetter(t)
			mockStockChecker := mocks.NewMockStockChecker(t)
			mockStockDecreaser := mocks.NewMockStockDecreaser(t)
			mockTxManager := mocks.NewMockTransactionManager(t)

			tt.setupMock(mockCreator, mockMenuGetter, mockStockChecker, mockStockDecreaser, mockTxManager)

			// Создаём юзкейс со всеми моками
			uc := NewCreateOrderUseCase(
				mockCreator,
				mockMenuGetter,
				mockStockChecker,
				mockStockDecreaser,
				mockTxManager,
			)

			order, err := uc.Execute(context.Background(), tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.input.ID, order.ID)
			}

			// Проверяем, что все ожидаемые вызовы были выполнены
			mockCreator.AssertExpectations(t)
			mockMenuGetter.AssertExpectations(t)
			mockStockChecker.AssertExpectations(t)
			mockStockDecreaser.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}
