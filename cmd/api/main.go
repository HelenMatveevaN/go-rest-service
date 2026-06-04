package main

import (
	"database/sql" // Добавили стандартный пакет для работы с SQL
	"log"
	"net/http"

	_ "github.com/lib/pq" // Стандартный драйвер PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4"

	api "go-rest-service/internal/delivery/http"
	"go-rest-service/internal/repository"
	"go-rest-service/internal/service"

)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/subscription_db?sslmode=disable"

	//1 - запуск миграций
	log.Println("Проверка и запуск миграций базы данных...")
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("Ошибка инициализации миграций: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}
	log.Println("Миграции успешно применены или база уже обновлена!")

	//2 - подключение к постгресу
	log.Println("Подключение к PostgreSQL...")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Ошибка при открытии базы данных: %v", err)
	}
	defer db.Close() // Закроем пул соединений при выходе из программы

	// Проверяем реальное соединение с базой (пингуем её)
	if err := db.Ping(); err != nil {
		log.Fatalf("База данных недоступна: %v", err)
	}
	log.Println("Успешное подключение к PostgreSQL!")

	//3 - сборка слоев
	// Инициализируем слой данных (Репозиторий)
	// Вместо repository.NewMemoryRepository() создаем наш новый репозиторий Postgres
	repo := repository.NewPostgresRepository(db)
	//repo := repository.NewMemoryRepository()

	// Передаем новый репозиторий в сервис. Всё остальное работает автоматически!
	svc := service.NewSubscriptionService(repo)
	// Инициализируем транспортный слой (Хэндлер) и передаем ему сервис
	handlers := api.NewHandler(svc)
	router := handlers.InitRoutes()

	//4 - запуск HTTP-сервера
	log.Println("Сервер подписок запускается на порту :8087...")
	if err := http.ListenAndServe(":8087", router); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}