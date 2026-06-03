//Привязываем HTTP-методы к конкретным функциям обработки

package http

import (
	"log" // Добавили импорт логов
	"net/http"
	
	"go-rest-service/internal/domain"
)

// Handler — управляет HTTP-запросами и хранит состояние (наша In-Memory БД)
type Handler struct {
	service			domain.SubscriptionService
}

func NewHandler(svc domain.SubscriptionService) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	//регистрируем маршруты для списка/создания операций и для операция по id
	log.Println("Регистрация пути 1: /api/v1/subscriptions")
	mux.HandleFunc("/api/v1/subscriptions", h.handleSubscriptions)

	log.Println("Регистрация пути 2: /api/v1/subscriptions/")
	mux.HandleFunc("/api/v1/subscriptions/", h.handleSubscriptionsWithID)

	return mux
}



