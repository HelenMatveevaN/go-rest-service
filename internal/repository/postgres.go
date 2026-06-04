package repository

import (
	"context"
	"fmt"
	"database/sql"
	
	"go-rest-service/internal/domain"
)

// PostgresRepository — новая структура, которая будет работать с реальной базой данных.
// Вместо карты памяти она хранит внутри себя пул соединений с Postgres (*sql.DB).
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository — конструктор, который принимает готовое подключение к базе
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Save записывает новую подписку в таблицу subscriptions
func (r *PostgresRepository) Save(ctx context.Context, sub domain.Subscription) error {
	// Пишем классический SQL-запрос на вставку строки.
	// Знак $1, $2 и т.д. — это плейсхолдеры. Они защищают наше приложение от SQL-инъекций.
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Выполняем запрос в базу данных.
	// Функция ExecContext принимает контекст, сам запрос и по порядку подставляет значения вместо $1, $2...
	_, err := r.db.ExecContext(
		ctx, 
		query, 
		sub.ID, 
		sub.ServiceName, 
		sub.Price, 
		sub.UserID, 
		sub.StartDate, 
		sub.EndDate, // Если указатель равен nil, драйвер pq автоматически запишет в базу значение NULL
		sub.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetByID находит одну конкретную подписку по её UUID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at 
		FROM subscriptions 
		WHERE id = $1
	`

	var sub domain.Subscription

	// QueryRowContext используется, когда мы гарантированно ждем только ОДНУ строку в ответ.
	// Метод .Scan() берет колонки из SELECT по порядку и записывает их по указателям в поля структуры sub.
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate, // Драйвер pq сам запишет сюда nil, если в базе лежит NULL, и выделит память, если там дата
		&sub.CreatedAt,
	)
	if err != nil {
		return domain.Subscription{}, err
	}

	return sub, nil
}

// List вытаскивает список подписок с фильтрацией по user_id
func (r *PostgresRepository) List(ctx context.Context, userID string) ([]domain.Subscription, error) {
	// Если userID передан, фильтруем по нему. Если нет — выводим вообще всё.
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at 
		FROM subscriptions 
		WHERE ($1 = '' OR user_id = $1::uuid)
		ORDER BY created_at DESC
	`

	// QueryContext используется для получения множества строк (списка)
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	// Обязательно закрываем rows в самом конце работы функции, чтобы освободить соединение с базой
	defer rows.Close()

	list := make([]domain.Subscription, 0)

	// Перебираем строки из базы одну за другой
	for rows.Next() {
		var sub domain.Subscription
		
		// Сканируем текущую строку в структуру sub
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		list = append(list, sub)
	}

	// Проверяем, не возникло ли ошибок во время итерации по строкам
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// Update полностью перезаписывает измененную модель подписки в базе данных
func (r *PostgresRepository) Update(ctx context.Context, sub domain.Subscription) error {
	query := `
		UPDATE subscriptions 
		SET service_name = $1, price = $2, end_date = $3 
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, sub.ServiceName, sub.Price, sub.EndDate, sub.ID)
	if err != nil {
		return err
	}

	return nil
}

// Delete физически удаляет запись о подписке по её UUID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Полезная проверка: узнаем, удалилось ли что-то на самом деле
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// Если база говорит, что удалено 0 строк, значит записи с таким ID не существовало
	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

// GetByFilters вытаскивает подписки пользователя с возможностью фильтрации по имени сервиса
func (r *PostgresRepository) GetByFilters(ctx context.Context, userID, serviceName string) ([]domain.Subscription, error) {
	// Пишем точечный SQL-запрос вместо старого перебора циклом for по всей карте памяти!
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at 
		FROM subscriptions 
		WHERE user_id = $1::uuid AND ($2 = '' OR service_name = $2)
	`

	rows, err := r.db.QueryContext(ctx, query, userID, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.Subscription, 0)
	for rows.Next() {
		var sub domain.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
