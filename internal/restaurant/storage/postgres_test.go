package storage

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/seWy-bit/GO-and-eat/internal/pkg/migrate"
	"github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RestaurantStorageSuite struct {
	suite.Suite
	ctx       context.Context
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
	storage   *PostgresStorage
}

// SetupSuite запускается 1 раз перед всеми тестами
func (s *RestaurantStorageSuite) SetupSuite() {
	s.ctx = context.Background()

	// Параметры:
	// - "postgres:15" — какой образ использовать
	// - postgres.WithDatabase("testdb") — создать БД с именем testdb
	// - postgres.WithUsername("testuser") — создать пользователя
	// - postgres.WithPassword("testpass") — пароль пользователя
	// - postgres.WithInitScripts() — можно выполнить SQL скрипты при старте

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

	// Получаем строку подключения к БД внутри контейнера
	connStr, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	// Создаём пул соединений
	poolConfig, err := pgxpool.ParseConfig(connStr)
	require.NoError(s.T(), err)

	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	pool, err := pgxpool.NewWithConfig(s.ctx, poolConfig)
	require.NoError(s.T(), err)

	err = pool.Ping(s.ctx)
	require.NoError(s.T(), err)

	s.pool = pool

	// Применяем миграции
	err = migrate.ApplyMigrationsForTest(s.ctx, s.pool, "../../../scripts/migrations/restaurant")
	require.NoError(s.T(), err, "failed to apply migrations")

	// Создаем storage
	s.storage = NewPostgresStorage(s.pool)
}

// TearDownSuite выполняется 1 раз после всех тестов
func (s *RestaurantStorageSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}

	if s.container != nil {
		err := s.container.Terminate(s.ctx)
		require.NoError(s.T(), err)
	}
}

// SetupTest выполняется перед каждым тестом
func (s *RestaurantStorageSuite) SetupTest() {
	err := migrate.CleanTables(s.ctx, s.pool)
	require.NoError(s.T(), err)
}

// createTestRestaurant создаёт тестовый ресторан
func (s *RestaurantStorageSuite) createTestRestaurant() domain.Restaurant {
	return domain.Restaurant{
		ID:        "rest-1",
		Name:      "ПиццаМания",
		Address:   "ул. Пушкина, 10",
		Phone:     "+7(999)123-45-67",
		CreatedAt: time.Now(),
	}
}

// createTestMenuItem создаёт тестовое блюдо
func (s *RestaurantStorageSuite) createTestMenuItem() domain.MenuItem {
	return domain.MenuItem{
		ID:           "pizza-1",
		RestaurantID: "rest-1",
		Name:         "Маргарита",
		Description:  "Томатный соус, моцарелла, базилик",
		Price:        59900,
		Stock:        10,
		Available:    true,
		CreatedAt:    time.Now(),
	}
}

// setupTestData создаёт ресторан и блюдо для тестов
func (s *RestaurantStorageSuite) setupTestData() error {
	restaurant := s.createTestRestaurant()
	if err := s.storage.CreateRestaurant(restaurant); err != nil {
		return err
	}

	item := s.createTestMenuItem()
	return s.storage.AddMenuItem(item)
}

func (s *RestaurantStorageSuite) TestCreateRestaurant() {
	restaurant := s.createTestRestaurant()

	err := s.storage.CreateRestaurant(restaurant)
	assert.NoError(s.T(), err)

	// Проверяем, что ресторан сохранился
	var count int
	err = s.pool.QueryRow(s.ctx, "SELECT COUNT(*) FROM restaurants WHERE id = 'rest-1'").Scan(&count)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)

	// Проверяем данные
	var name, address, phone string
	err = s.pool.QueryRow(s.ctx,
		"SELECT name, address, phone FROM restaurants WHERE id = 'rest-1'",
	).Scan(&name, &address, &phone)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "ПиццаМания", name)
	assert.Equal(s.T(), "ул. Пушкина, 10", address)
	assert.Equal(s.T(), "+7(999)123-45-67", phone)
}

func (s *RestaurantStorageSuite) TestCreateRestaurant_Duplicate() {
	restaurant := s.createTestRestaurant()

	err := s.storage.CreateRestaurant(restaurant)
	assert.NoError(s.T(), err)

	err = s.storage.CreateRestaurant(restaurant)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "already exists")
}

func (s *RestaurantStorageSuite) TestAddMenuItem() {
	// Сначала создаём ресторан
	restaurant := s.createTestRestaurant()
	err := s.storage.CreateRestaurant(restaurant)
	require.NoError(s.T(), err)

	// Добавляем блюдо
	item := s.createTestMenuItem()
	err = s.storage.AddMenuItem(item)
	assert.NoError(s.T(), err)

	// Проверяем, что блюдо сохранилось
	var count int
	err = s.pool.QueryRow(s.ctx, "SELECT COUNT(*) FROM menu_items WHERE id = 'pizza-1'").Scan(&count)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
}

func (s *RestaurantStorageSuite) TestAddMenuItem_RestaurantNotFound() {
	item := s.createTestMenuItem()
	item.RestaurantID = "rest-not-exist"

	err := s.storage.AddMenuItem(item)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "restaurant not found")
}

func (s *RestaurantStorageSuite) TestGetMenu() {
	// Создаём ресторан и добавляем блюда
	restaurant := s.createTestRestaurant()
	err := s.storage.CreateRestaurant(restaurant)
	require.NoError(s.T(), err)

	item1 := domain.MenuItem{
		ID:           "pizza-1",
		RestaurantID: "rest-1",
		Name:         "Маргарита",
		Description:  "Томатный соус, моцарелла",
		Price:        59900,
		Stock:        10,
		Available:    true,
		CreatedAt:    time.Now(),
	}
	err = s.storage.AddMenuItem(item1)
	require.NoError(s.T(), err)

	// Небольшая задержка, чтобы created_at точно отличался
	time.Sleep(10 * time.Millisecond)

	item2 := domain.MenuItem{
		ID:           "pizza-2",
		RestaurantID: "rest-1",
		Name:         "Пепперони",
		Description:  "Острая колбаса, сыр",
		Price:        69900,
		Stock:        5,
		Available:    true,
		CreatedAt:    time.Now(),
	}
	err = s.storage.AddMenuItem(item2)
	require.NoError(s.T(), err)

	// Получаем меню
	menu, err := s.storage.GetMenu("rest-1")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), menu, 2)

	// Проверяем, что оба блюда есть в меню (независимо от порядка)
	// Создаём map для быстрой проверки
	menuMap := make(map[string]domain.MenuItem)
	for _, item := range menu {
		menuMap[item.ID] = item
	}

	// Проверяем, что pizza-1 есть
	pizza1, exists := menuMap["pizza-1"]
	assert.True(s.T(), exists, "pizza-1 should be in menu")
	assert.Equal(s.T(), "Маргарита", pizza1.Name)
	assert.Equal(s.T(), int64(59900), pizza1.Price)

	// Проверяем, что pizza-2 есть
	pizza2, exists := menuMap["pizza-2"]
	assert.True(s.T(), exists, "pizza-2 should be in menu")
	assert.Equal(s.T(), "Пепперони", pizza2.Name)
	assert.Equal(s.T(), int64(69900), pizza2.Price)
}

func (s *RestaurantStorageSuite) TestGetMenu_Empty() {
	restaurant := s.createTestRestaurant()
	err := s.storage.CreateRestaurant(restaurant)
	require.NoError(s.T(), err)

	menu, err := s.storage.GetMenu("rest-1")
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), menu) // Должен вернуть пустой слайс, а не nil
}

func (s *RestaurantStorageSuite) TestGetMenu_RestaurantNotFound() {
	menu, err := s.storage.GetMenu("rest-not-exist")
	assert.NoError(s.T(), err) // Не должно быть ошибки
	assert.Empty(s.T(), menu)  // Просто пустое меню
}

func (s *RestaurantStorageSuite) TestDecreaseStockWithTx() {
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	err = s.storage.DecreaseStockWithTx(s.ctx, tx, "rest-1", "pizza-1", 3)
	assert.NoError(s.T(), err)

	var stock int
	var available bool
	err = tx.QueryRow(s.ctx,
		"SELECT stock, available FROM menu_items WHERE id = 'pizza-1'").Scan(&stock, &available)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 7, stock)
	assert.True(s.T(), available)

	err = tx.Rollback(s.ctx)
	assert.NoError(s.T(), err)

	err = s.pool.QueryRow(s.ctx, "SELECT stock FROM menu_items WHERE id = 'pizza-1'").Scan(&stock)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 10, stock)
}

func (s *RestaurantStorageSuite) TestDecreaseStockWithTx_NotEnough() {
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	err = s.storage.DecreaseStockWithTx(s.ctx, tx, "rest-1", "pizza-1", 15)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not enough stock")
}

func (s *RestaurantStorageSuite) TestDecreaseStockWithTx_ItemNotFound() {
	// Подготовка
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	err = s.storage.DecreaseStockWithTx(s.ctx, tx, "rest-1", "pizza-not-exist", 5)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "menu item not found")
}

func (s *RestaurantStorageSuite) TestCheckAvailabilityWithTx() {
	// Подготовка
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	checkItems := []struct {
		ID       string
		Quantity int
	}{
		{ID: "pizza-1", Quantity: 5},
	}

	available, err := s.storage.CheckAvailabilityWithTx(s.ctx, tx, "rest-1", checkItems)
	assert.NoError(s.T(), err)
	assert.True(s.T(), available)
}

func (s *RestaurantStorageSuite) TestCheckAvailabilityWithTx_NotEnough() {
	// Подготовка
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	checkItems := []struct {
		ID       string
		Quantity int
	}{
		{ID: "pizza-1", Quantity: 15}, // 15 > 10
	}

	available, err := s.storage.CheckAvailabilityWithTx(s.ctx, tx, "rest-1", checkItems)
	assert.NoError(s.T(), err)
	assert.False(s.T(), available)
}

func (s *RestaurantStorageSuite) TestCheckAvailabilityWithTx_ItemNotFound() {
	// Подготовка
	err := s.setupTestData()
	require.NoError(s.T(), err)

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	checkItems := []struct {
		ID       string
		Quantity int
	}{
		{ID: "pizza-not-exist", Quantity: 5},
	}

	available, err := s.storage.CheckAvailabilityWithTx(s.ctx, tx, "rest-1", checkItems)
	assert.NoError(s.T(), err)
	assert.False(s.T(), available) // Блюдо не найдено → недоступно
}

func (s *RestaurantStorageSuite) TestCheckAvailabilityWithTx_MultipleItems() {
	// Подготовка
	restaurant := s.createTestRestaurant()
	err := s.storage.CreateRestaurant(restaurant)
	require.NoError(s.T(), err)

	items := []domain.MenuItem{
		{
			ID:           "pizza-1",
			RestaurantID: "rest-1",
			Name:         "Маргарита",
			Price:        59900,
			Stock:        10,
			Available:    true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           "pizza-2",
			RestaurantID: "rest-1",
			Name:         "Пепперони",
			Price:        69900,
			Stock:        5,
			Available:    true,
			CreatedAt:    time.Now(),
		},
	}

	for _, item := range items {
		err = s.storage.AddMenuItem(item)
		require.NoError(s.T(), err)
	}

	tx, err := s.pool.Begin(s.ctx)
	require.NoError(s.T(), err)
	defer tx.Rollback(s.ctx)

	checkItems := []struct {
		ID       string
		Quantity int
	}{
		{ID: "pizza-1", Quantity: 2},
		{ID: "pizza-2", Quantity: 3},
	}

	available, err := s.storage.CheckAvailabilityWithTx(s.ctx, tx, "rest-1", checkItems)
	assert.NoError(s.T(), err)
	assert.True(s.T(), available)

	// Тест: одно блюдо недоступно
	checkItems2 := []struct {
		ID       string
		Quantity int
	}{
		{ID: "pizza-1", Quantity: 2},
		{ID: "pizza-2", Quantity: 10}, // 10 > 5
	}

	available2, err := s.storage.CheckAvailabilityWithTx(s.ctx, tx, "rest-1", checkItems2)
	assert.NoError(s.T(), err)
	assert.False(s.T(), available2)
}

func TestRestaurantStorageSuite(t *testing.T) {
	suite.Run(t, new(RestaurantStorageSuite))
}
