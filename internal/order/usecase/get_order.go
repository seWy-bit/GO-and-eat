package usecase

import (
	"context"
	"errors"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
)

type GetOrderUseCase struct {
	orderGetter OrderGetter
}

func NewGetOrderUseCase(orderGetter OrderGetter) *GetOrderUseCase {
	return &GetOrderUseCase{
		orderGetter: orderGetter,
	}
}

func (uc *GetOrderUseCase) Execute(ctx context.Context, id string) (*domain.Order, error) {
	if id == "" {
		return nil, errors.New("order id is required")
	}

	order, err := uc.orderGetter.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
