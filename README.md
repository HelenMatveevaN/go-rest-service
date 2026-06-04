# Subscription REST API Service

REST API сервис для учета и аналитики подписок пользователей (например, Яндекс Плюс, Иви и др.) с возможностью расчета суммарной стоимости затрат за выбранный период.

Проект спроектирован по канонам **Чистой Архитектуры (Clean Architecture)** с четким разделением на слои и зависимостями, направленными внутрь.

## 🛠 Технологический стек

* **Language:** Go 1.21 (стандартная библиотека `net/http`)
* **Database:** PostgreSQL 15
* **Database Migrations:** `golang-migrate` (автоматический накат схем при старте)
* **Containerization:** Docker, Docker Compose
* **API Documentation:** OpenAPI 3.0 (Swagger YAML)
* **Logging:** Сквозное структурированное логирование (HTTP -> Service -> Database)

## 🏗 Архитектура проекта

```text
internal/
├── domain/       # Ядро: модели, DTO и интерфейсы слоев (бизнес-контракты)
├── delivery/     # Транспортный слой: HTTP-хэндлеры (чистый net/http)
├── service/      # Бизнес-логика: валидация, парсинг, калькулятор пересечения дат
└── repository/   # Слой данных: SQL-запросы к PostgreSQL
```

## 🚀 Быстрый запуск (Docker Compose)

Для запуска всей инфраструктуры (сервис + база данных + автоматические миграции) необходим только установленный Docker Desktop.

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com
   cd go-rest-service
   ```

2. Запустите контейнеры одной командой:
   ```bash
   docker compose up --build
   ```

Сервер автоматически применит SQL-миграции, создаст таблицы, индексы и станет доступен на порту `:8087`.

## 📋 Примеры API запросов

### 1. Создание подписки
`POST /api/v1/subscriptions`
```bash
curl -X POST http://localhost:8087/api/v1/subscriptions \
-H "Content-Type: application/json" \
-d '{"service_name": "Yandex Plus", "price": 400, "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba", "start_date": "07-2025"}'
```

### 2. Расчет стоимости подписок за период (с учетом дат активности)
`GET /api/v1/subscriptions/total-price`
```bash
curl -i "http://localhost:8087/api/v1/subscriptions/total-price?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&from=07-2025&to=09-2025"
```

*Спецификация всех ручек (CRUDL) доступна в файле `swagger.yaml`.*
