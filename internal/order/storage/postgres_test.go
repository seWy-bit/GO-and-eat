package storage

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	"github.com/seWy-bit/GO-and-eat/internal/pkg/migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	restaurantDomain "github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
	restaurantStorage "github.com/seWy-bit/GO-and-eat/internal/restaurant/storage"
)

type OrderStorageSuite struct {
	suite.Suite
	ctx               context.Context
	container         *postgres.PostgresContainer
	pool              *pgxpool.Pool
	orderStorage      *PostgresOrderStorage
	restaurantStorage *restaurantStorage.PostgresStorage
}

func (s *OrderStorageSuite) SetupSuite() {
	s.ctx = context.Background()

	pgContainer, err := postgres.Run(s.ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(s.T(), err, "failed to start postgres container")
	s.container = pgContainer

	connStr, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	require.NoError(s.T(), err)

	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	pool, err := pgxpool.NewWithConfig(s.ctx, poolConfig)
	require.NoError(s.T(), err)

	err = pool.Ping(s.ctx)
	require.NoError(s.T(), err)

	s.pool = pool

	err = migrate.ApplyMigrationsForTest(s.ctx, s.pool, "../../../scripts/migrations/restaurant")
	require.NoError(s.T(), err, "failed to apply migrations")

	s.orderStorage = NewPostgresOrderStorage(pool)
	s.restaurantStorage = restaurantStorage.NewPostgresStorage(pool)
}

func (s *OrderStorageSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}

	if s.container != nil {
		err := s.container.Terminate(s.ctx)
		require.NoError(s.T(), err)
	}
}

func (s *OrderStorageSuite) SetupTest() {
	err := migrate.CleanTables(s.ctx, s.pool)
	require.NoError(s.T(), err)
}

// createTestRestaurant создаёт тестовый ресторан
func (s *OrderStorageSuite) createTestRestaurant() error {
	restaurant := restaurantDomain.Restaurant{
		ID:        "rest-1",
		Name:      "ПиццаМания",
		Address:   "ул. Пушкина, 10",
		Phone:     "+7(999)123-45-67",
		CreatedAt: time.Now(),
	}
	return s.restaurantStorage.CreateRestaurant(restaurant)
}

// createTestMenuItem создаёт тестовое блюдо
func (s *OrderStorageSuite) createTestMenuItem() error {
	item := restaurantDomain.MenuItem{
		ID:           "pizza-1",
		RestaurantID: "rest-1",
		Name:         "Маргарита",
		Description:  "Томатный соус, моцарелла, базилик",
		Price:        59900,
		Stock:        10,
		Available:    true,
		CreatedAt:    time.Now(),
	}
	return s.restaurantStorage.AddMenuItem(item)
}

// setupTestData подготавливает данные для тестов
func (s *OrderStorageSuite) setupTestData() error {
	if err := s.createTestRestaurant(); err != nil {
		return err
	}
	return s.createTestMenuItem()
}

// createTestOrder создаёт тестовый заказ
func (s *OrderStorageSuite) createTestOrder() domain.Order {
	now := time.Now()
	return domain.Order{
		ID:           "order-1",
		UserID:       "user-1",
		RestaurantID: "rest-1",
		Status:       domain.OrderStatusCreated,
		TotalAmount:  119800,
		Items: []domain.OrderItem{
			{
				MenuItemID: "pizza-1",
				Name:       "Маргарита",
				Quantity:   2,
				Price:      59900,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s *OrderStorageSuite) TestCreateOrderWithTx() {
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	order := s.createTestOrder()
	err = s.orderStorage.CreateOrderWithTx(s.ctx, tx, order)
	assert.NoError(s.T(), err)

	var count int
	err = tx.QueryRow(s.ctx, "SELECT COUNT(*) FROM orders WHERE id = 'order-1'").Scan(&count)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)

	var itemsCount int
	err = tx.QueryRow(s.ctx, "SELECT COUNT(*) FROM order_items WHERE order_id = 'order-1'").Scan(&itemsCount)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, itemsCount)
}

func (s *OrderStorageSuite) TestCreateOrderWithTx_Rollback() {
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)

	order := s.createTestOrder()
	err = s.orderStorage.CreateOrderWithTx(s.ctx, tx, order)
	assert.NoError(s.T(), err)

	err = tx.Rollback(s.ctx)
	assert.NoError(s.T(), err)

	var count int
	err = s.pool.QueryRow(s.ctx, "SELECT COUNT(*) FROM orders WHERE id = 'order-1'").Scan(&count)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *OrderStorageSuite) TestGetOrder() {
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	order := s.createTestOrder()
	err = s.orderStorage.CreateOrderWithTx(s.ctx, tx, order)
	require.NoError(s.T(), err)

	err = tx.Commit(s.ctx)
	require.NoError(s.T(), err)

	received, err := s.orderStorage.GetOrder(s.ctx, "order-1")
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), "order-1", received.ID)
	assert.Equal(s.T(), "user-1", received.UserID)
	assert.Equal(s.T(), "rest-1", received.RestaurantID)
	assert.Equal(s.T(), domain.OrderStatusCreated, received.Status)
	assert.Equal(s.T(), int64(119800), received.TotalAmount)

	assert.Len(s.T(), received.Items, 1)
	assert.Equal(s.T(), "pizza-1", received.Items[0].MenuItemID)
	assert.Equal(s.T(), "Маргарита", received.Items[0].Name)
	assert.Equal(s.T(), 2, received.Items[0].Quantity)
	assert.Equal(s.T(), int64(59900), received.Items[0].Price)
}

func (s *OrderStorageSuite) TestGetOrder_NotFound() {
	_, err := s.orderStorage.GetOrder(s.ctx, "order-not-exist")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "order not found")
}

func (s *OrderStorageSuite) TestUpdateOrderStatus() {
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)

	order := s.createTestOrder()
	err = s.orderStorage.CreateOrderWithTx(s.ctx, tx, order)
	require.NoError(s.T(), err)

	err = tx.Commit(s.ctx)
	require.NoError(s.T(), err)

	err = s.orderStorage.UpdateOrderStatus(s.ctx, "order-1", domain.OrderStatusConfirmed)
	assert.NoError(s.T(), err)

	var status string
	err = s.pool.QueryRow(s.ctx, "SELECT status FROM orders WHERE id = 'order-1'").Scan(&status)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "confirmed", status)
}

func (s *OrderStorageSuite) TestUpdateOrderStatus_NotFound() {
	err := s.orderStorage.UpdateOrderStatus(s.ctx, "order-not-exist", domain.OrderStatusConfirmed)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "order not found")
}

func (s *OrderStorageSuite) TestGetOrdersByUser() {
	// Подготовка данных
	err := s.setupTestData()
	require.NoError(s.T(), err)

	// Создаём несколько заказов
	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)

	order1 := s.createTestOrder()
	order1.ID = "order-1"
	err = s.orderStorage.CreateOrderWithTx(s.ctx, tx, order1)
	require.NoError(s.T(), err)

	order2 := s.createTestOrder()
	order2.ID = "order-2"
	err = s.orderStorage.CreateOrderWithTx(s.ctx, tx, order2)
	require.NoError(s.T(), err)

	err = tx.Commit(s.ctx)
	require.NoError(s.T(), err)

	// Получаем заказы пользователя
	orders, err := s.orderStorage.GetOrdersByUser(s.ctx, "user-1")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), orders, 2)

	// Проверяем, что оба заказа принадлежат пользователю
	for _, order := range orders {
		assert.Equal(s.T(), "user-1", order.UserID)
	}
}

func (s *OrderStorageSuite) TestGetOrdersByUser_Empty() {
	orders, err := s.orderStorage.GetOrdersByUser(s.ctx, "user-no-orders")
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), orders) // Должен вернуть пустой слайс, а не nil
}

func TestOrderStorageSuite(t *testing.T) {
	suite.Run(t, new(OrderStorageSuite))
}
