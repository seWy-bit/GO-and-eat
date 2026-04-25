package usecase

import (
	"context"
	"errors"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	"github.com/seWy-bit/GO-and-eat/internal/order/storage"
)

type GetUserOrdersUseCase struct {
	orderStorage *storage.PostgresOrderStorage
}

func NewGetUserOrdersUseCase(orderStorage *storage.PostgresOrderStorage) *GetUserOrdersUseCase {
	return &GetUserOrdersUseCase{
		orderStorage: orderStorage,
	}
}

func (uc *GetUserOrdersUseCase) Execute(ctx context.Context, userID string) ([]domain.Order, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	orders, err := uc.orderStorage.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
