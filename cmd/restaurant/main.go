package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderHandler "github.com/seWy-bit/GO-and-eat/internal/order/handler"
	orderStorage "github.com/seWy-bit/GO-and-eat/internal/order/storage"
	orderUsecase "github.com/seWy-bit/GO-and-eat/internal/order/usecase"
	"github.com/seWy-bit/GO-and-eat/internal/pkg/config"
	"github.com/seWy-bit/GO-and-eat/internal/pkg/database"
	"github.com/seWy-bit/GO-and-eat/internal/pkg/migrate"

	restaurantHandler "github.com/seWy-bit/GO-and-eat/internal/restaurant/handler"
	restaurantStorage "github.com/seWy-bit/GO-and-eat/internal/restaurant/storage"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	dbCfg := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	log.Println("Connected to database successfully")

	migrationRunner := migrate.NewRunner(pool.Pool)
	if err := migrationRunner.Up(context.Background(), "scripts/migrations/restaurant"); err != nil {
		log.Fatal("failed to run migrations:", err)
	}
	log.Println("Migrations applied successfully")

	restaurantStore := restaurantStorage.NewPostgresStorage(pool.Pool)
	restaurantHandlers := restaurantHandler.NewRestaurantHandler(restaurantStore)

	orderStore := orderStorage.NewPostgresOrderStorage(pool.Pool)
	createOrderUseCase := orderUsecase.NewCreateOrderUseCase(orderStore, restaurantStore, pool.Pool)
	getOrderUseCase := orderUsecase.NewGetOrderUseCase(orderStore)
	updateOrderStatusUseCase := orderUsecase.NewUpdateOrderStatusUseCase(orderStore)
	getUserOrdersUseCase := orderUsecase.NewGetUserOrdersUseCase(orderStore)

	orderHandlers := orderHandler.NewOrderHandler(
		createOrderUseCase,
		getOrderUseCase,
		updateOrderStatusUseCase,
		getUserOrdersUseCase,
	)

	// Restaurant endpoints
	http.HandleFunc("POST /restaurants", restaurantHandlers.CreateRestaurant)
	http.HandleFunc("GET /restaurants/{id}/menu", restaurantHandlers.GetMenu)
	http.HandleFunc("POST /restaurants/{id}/menu", restaurantHandlers.AddMenuItem)

	// Order endpoints
	http.HandleFunc("POST /orders", orderHandlers.CreateOrder)
	http.HandleFunc("GET /orders/{id}", orderHandlers.GetOrder)
	http.HandleFunc("PATCH /orders/{id}/status", orderHandlers.UpdateOrderStatus)
	http.HandleFunc("GET /orders/user/{user_id}", orderHandlers.GetUserOrders)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf(" 	Server starting on port %d", cfg.Server.Port)
		log.Printf(" 	REST endpoints:")
		log.Printf("   	POST   /restaurants")
		log.Printf("   	GET    /restaurants/{id}/menu")
		log.Printf("   	POST   /restaurants/{id}/menu")
		log.Printf("   	POST   /orders")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
