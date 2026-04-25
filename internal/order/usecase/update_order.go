package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	"github.com/seWy-bit/GO-and-eat/internal/order/storage"
)

type UpdateOrderStatusUseCase struct {
	orderStorage *storage.PostgresOrderStorage
}

func NewUpdateOrderStatusUseCase(orderStorage *storage.PostgresOrderStorage) *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{
		orderStorage: orderStorage,
	}
}

func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, id string, newStatus domain.OrderStatus) error {
	if id == "" {
		return errors.New("order id is required")
	}

	if newStatus == "" {
		return errors.New("status is required")
	}

	order, err := uc.orderStorage.GetOrder(ctx, id)
	if err != nil {
		return err
	}

	if !order.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition: %s -> %s", order.Status, newStatus)
	}

	if err = uc.orderStorage.UpdateOrderStatus(ctx, id, newStatus); err != nil {
		return err
	}

	return nil
}
