package main

import (
	"fmt"
	"log"
	"net/http"

	orderHandler "github.com/seWy-bit/GO-and-eat/internal/order/handler"
	orderStorage "github.com/seWy-bit/GO-and-eat/internal/order/storage"
	orderUsecase "github.com/seWy-bit/GO-and-eat/internal/order/usecase"

	restaurantHandler "github.com/seWy-bit/GO-and-eat/internal/restaurant/handler"
	restaurantStorage "github.com/seWy-bit/GO-and-eat/internal/restaurant/storage"
)

func main() {
	restaurantStore := restaurantStorage.NewMemoryStorage()
	restaurantHandlers := restaurantHandler.NewRestaurantHandler(restaurantStore)

	orderStore := orderStorage.NewMemoryOrderStorage()
	createOrderUseCase := orderUsecase.NewCreateOrderUseCase(orderStore, restaurantStore)
	orderHandlers := orderHandler.NewOrderHandler(createOrderUseCase)

	http.HandleFunc("POST /restaurants", restaurantHandlers.CreateRestaurant)
	http.HandleFunc("GET /restaurants/{id}/menu", restaurantHandlers.GetMenu)
	http.HandleFunc("POST /restaurants/{id}/menu", restaurantHandlers.AddMenuItem)

	http.HandleFunc("POST /orders", orderHandlers.CreateOrder)

	port := 8081
	fmt.Printf("Сервер запущен на порту %d\n", port)
	fmt.Println("\nДоступные эндпоинты:")
	fmt.Println("  POST   /restaurants")
	fmt.Println("  GET    /restaurants/{id}/menu")
	fmt.Println("  POST   /restaurants/{id}/menu")
	fmt.Println("  POST   /orders")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
