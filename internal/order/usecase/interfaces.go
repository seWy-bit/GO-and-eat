package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	restaurantDomain "github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
)

type OrderGetter interface {
	GetOrder(ctx context.Context, id string) (domain.Order, error)
}

type OrderCreator interface {
	CreateOrderWithTx(ctx context.Context, tx pgx.Tx, order domain.Order) error
}

type OrderStatusUpdater interface {
	UpdateOrderStatus(ctx context.Context, id string, newStatus domain.OrderStatus) error
}

type UserOrdersGetter interface {
	GetOrdersByUser(ctx context.Context, userID string) ([]domain.Order, error)
}

type MenuGetter interface {
	GetMenu(restaurantID string) ([]restaurantDomain.MenuItem, error)
}

type StockChecker interface {
	CheckAvailabilityWithTx(ctx context.Context, tx pgx.Tx, restaurantID string, items []struct {
		ID       string
		Quantity int
	}) (bool, error)
}

type StockDecreaser interface {
	DecreaseStockWithTx(ctx context.Context, tx pgx.Tx, restaurantID, menuItemID string, quantity int) error
}

type TransactionManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Commit(tx pgx.Tx) error
	Rollback(tx pgx.Tx) error
}
