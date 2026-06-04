package main

import (
	"bufio" // для построчного чтения файла
	"database/sql" // Добавили стандартный пакет для работы с SQL
	"log"
	"net/http"
	"os" // для работы с переменными окружения
	"strings"

	_ "github.com/lib/pq" // Стандартный драйвер PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4"

	api "go-rest-service/internal/delivery/http"
	"go-rest-service/internal/repository"
	"go-rest-service/internal/service"

)

// initEnv — функция для ручного парсинга .env файла
func initEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Println("[WARN] Файл .env не найден, используются системные переменные окружения")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Разбиваем строку по первому знаку "="
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Загружаем переменную в окружение приложения
		os.Setenv(key, value)
	}
	log.Println("[INFO] Конфигурация из .env файла успешно загружена")
}


func main() {
	//1 - Загружаем настройки из конфигурационного файла
	initEnv()

	//2 - Достаем настройки из окружения с дефолтными значениями на случай, если файла нет
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("[FATAL] Переменная окружения DB_URL не задана")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = ":8087" // Дефолтный порт, если забыли указать
	}

	//3 - запуск миграций
	log.Println("Проверка и запуск миграций базы данных...")
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("Ошибка инициализации миграций: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}
	log.Println("Миграции успешно применены или база уже обновлена!")

	//4 - подключение к постгресу
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

	//5 - сборка слоев
	// Инициализируем слой данных (Репозиторий)
	// Вместо repository.NewMemoryRepository() создаем наш новый репозиторий Postgres
	repo := repository.NewPostgresRepository(db)
	//repo := repository.NewMemoryRepository()

	// Передаем новый репозиторий в сервис. Всё остальное работает автоматически!
	svc := service.NewSubscriptionService(repo)
	// Инициализируем транспортный слой (Хэндлер) и передаем ему сервис
	handlers := api.NewHandler(svc)
	router := handlers.InitRoutes()

	//6 - запуск HTTP-сервера
	log.Println("Сервер подписок запускается на порту :8087...")
	if err := http.ListenAndServe(":8087", router); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}