package usecase

import (
	"context"
	"errors"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	"github.com/seWy-bit/GO-and-eat/internal/order/storage"
)

type GetOrderUseCase struct {
	orderStorage *storage.PostgresOrderStorage
}

func NewGetOrderUseCase(orderStorage *storage.PostgresOrderStorage) *GetOrderUseCase {
	return &GetOrderUseCase{
		orderStorage: orderStorage,
	}
}

func (uc *GetOrderUseCase) Execute(ctx context.Context, id string) (*domain.Order, error) {
	if id == "" {
		return nil, errors.New("order id is required")
	}

	order, err := uc.orderStorage.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
