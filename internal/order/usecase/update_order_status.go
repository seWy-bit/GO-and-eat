package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
)

type UpdateOrderStatusUseCase struct {
	orderGetter  OrderGetter
	orderUpdater OrderStatusUpdater
}

func NewUpdateOrderStatusUseCase(orderGetter OrderGetter, orderUpdater OrderStatusUpdater) *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{
		orderGetter:  orderGetter,
		orderUpdater: orderUpdater,
	}
}

func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, id string, newStatus domain.OrderStatus) error {
	if id == "" {
		return errors.New("order id is required")
	}

	if newStatus == "" {
		return errors.New("status is required")
	}

	order, err := uc.orderGetter.GetOrder(ctx, id)
	if err != nil {
		return err
	}

	if !order.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition: %s -> %s", order.Status, newStatus)
	}

	if err = uc.orderUpdater.UpdateOrderStatus(ctx, id, newStatus); err != nil {
		return err
	}

	return nil
}
