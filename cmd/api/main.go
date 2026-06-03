package main

import (
	"log"
	"net/http"

	api "go-rest-service/internal/delivery/http"
	"go-rest-service/internal/repository"
	"go-rest-service/internal/service"

)

func main() {
	// 1. Инициализируем слой данных (Репозиторий)
	repo := repository.NewMemoryRepository()

	// 2. Инициализируем слой бизнес-логики (Сервис) и передаем ему репозиторий
	svc := service.NewSubscriptionService(repo)

	// 3. Инициализируем транспортный слой (Хэндлер) и передаем ему сервис
	handlers := api.NewHandler(svc)
	router := handlers.InitRoutes()

	//run http-serv on port 8087
	log.Println("Сервер подписок запускается на порту :8087...")
	if err := http.ListenAndServe(":8087", router); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}