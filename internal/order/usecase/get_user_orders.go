package usecase

import (
	"context"
	"errors"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
)

type GetUserOrdersUseCase struct {
	userOrdersGetter UserOrdersGetter
}

func NewGetUserOrdersUseCase(userOrdersGetter UserOrdersGetter) *GetUserOrdersUseCase {
	return &GetUserOrdersUseCase{
		userOrdersGetter: userOrdersGetter,
	}
}

func (uc *GetUserOrdersUseCase) Execute(ctx context.Context, userID string) ([]domain.Order, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	orders, err := uc.userOrdersGetter.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
