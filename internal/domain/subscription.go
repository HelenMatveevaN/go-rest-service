//модели данных и интерф.бизнес-логики
//слой описания данных и правил

package domain

import (
	"context"
	"errors"
	"time"
)

//проверка валидности дат подписки
var ErrInvalidDates = errors.New("end_date cannot be before start_date")

//подписки
type Subscription struct {
	ID				string		`json:"id"`
	ServiceName		string		`json:"service_name"`
	Price			int			`json:"price"`				// Стоимость в целых рублях
	UserID			string		`json:"user_id"`			// Строка формата UUID
	StartDate		time.Time	`json:"start_date"` 		//месяц и год начала
	EndDate			*time.Time	`json:"end_date,omitempty"`	//м.б.пустым
	CreatedAt		time.Time	`json:"created_at"`
}

//входные данные для POST /api/v1/subscription
//определяем структуру д-х CreateInput 
//(парсинг и валидация json от клиента в HTTP-q POST - создание подписки)
type CreateInput struct {
	ServiceName		string		`json:"service_name" binding:"required"`  //обяз.поле
	Price			int			`json:"price" binding:"required",gt=0` //gt — greater than
	UserID			string		`json:"user_id" binding:"required",uuid4`
	StartDate		string		`json:"start_date" binding:"required"`	  // MM-YYYY
	EndDate			*string		`json:"end_date" binding:"omitempty"`	  // MM-YYYY
}

//вх.д-ые для PATCH /api/v1/subscription/:id
type UpdateInput struct {
	ServiceName		*string		`json:"service_name" binding:"omitempty`
	Price			*int		`json:"price" binding:"omitempty",gt=0` 
	EndDate			**string	`json:"end_date" binding:"omitempty"`	
}

//структура - ответ ручки для подсчета стоимости
type GetTotalCostOutput struct {
	TotalCost int `json:"total_cost"`
}

//КОНТРАКТЫ:

//описание бизнес-логики для CRUDL
//прием сырых данных от польз-ля и проверка бизнес-правил
type SubscriptionService interface {
	Create(ctx context.Context, input CreateInput) (Subscription, error)
	GetByID(ctx context.Context, id string) (Subscription, error)
	Update(ctx context.Context, id string, input UpdateInput) (Subscription, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, userID string) ([]Subscription, error)
	//in-фильтры,out-структура ответа
	GetTotalCost(ctx context.Context, userID, ServiceName, fromDate, toDate string) (GetTotalCostOutput, error)
}

//описание методов работы с хранилищем (БД)
//задача - сохранить или достать готовую сущность из БД
type SubscriptionRepository interface {
	Save(ctx context.Context, sub Subscription) error
	GetByID(ctx context.Context, id string) (Subscription, error)
	Update(ctx context.Context, sub Subscription) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, userID string) ([]Subscription, error)
	//get subscrip-s by userID, filter by ServName
	GetByFilters(ctx context.Context, userID, ServiceName string) ([]Subscription, error)
}

