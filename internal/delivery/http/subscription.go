//логика чтения запросов и отправки ответов (валидация, парсинг JSON/Query)

package http

import (
	"encoding/json"
	"log"
	"strings"
	"net/http"

	"go-rest-service/internal/domain"
)

// handleSubscriptions распределяет запросы к общему списку: POST (Create) и GET (List)
func (h *Handler) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Логируем метод и путь для каждого входящего запроса на этот эндпоинт
	log.Printf("[INFO] Входящий запрос: %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodPost:
		h.createSubscription(w, r)
	case http.MethodGet:
		h.listSubscriptions(w, r)
	default:
		log.Printf("[WARN] Метод %s не разрешен для пути %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleSubscriptionsWithID распределяет запросы по ID подписки: 
//GET (Read), PATCH (Update), DELETE (Delete)
func (h *Handler) handleSubscriptionsWithID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("[INFO] Входящий запрос: %s %s", r.Method, r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")

	//перехыватываем ручку подсчета стоимости
	if id == "total-price" {
		if r.Method == http.MethodGet {
			h.getTotalCost(w, r)
			return
		}
		log.Printf("[WARN] Метод %s не разрешен для /total-price", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if id == "" {
		h.handleSubscriptions(w, r)
		return
	}
	
	switch r.Method {
	case http.MethodGet:
		h.getSubscription(w, r)
	case http.MethodPatch:
		h.updateSubscription(w, r)
	case http.MethodDelete:
		h.deleteSubscription(w, r)
	default:
		log.Printf("[WARN] Метод %s не разрешен для путей с ID", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// C - CREATE: POST /api/v1/subscriptions
func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("[ERROR] Не удалось распарсить JSON запроса: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	if input.ServiceName == "" || input.Price <= 0 || len(input.UserID) != 36 || input.StartDate == "" {
		log.Printf("[WARN] Ошибка валидации полей при создании подписки для user_id: %s", input.UserID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed: missing fields or invalid user_id format"})
		return
	}

	newSub, err := h.service.Create(r.Context(), input)
	if err != nil {
		log.Printf("[ERROR] Ошибка выполнения бизнес-логики создания подписки: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("[INFO] Подписка успешно создана с ID: %s для пользователя: %s", newSub.ID, newSub.UserID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newSub)
}

// R - READ: GET /api/v1/subscriptions/{id}
func (h *Handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")

	newSub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("[WARN] Подписка с ID %s не найдена в базе данных", id)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "subscription not found"})
		return
	}

	log.Printf("[INFO] Успешно получена подписка с ID: %s", id)
	json.NewEncoder(w).Encode(newSub)
}

// U - UPDATE: PATCH /api/v1/subscriptions/{id}
func (h *Handler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")

	var input domain.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("[ERROR] Не удалось распарсить JSON для обновления подписки %s: %v", id, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	updatedSub, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		log.Printf("[ERROR] Не удалось обновить подписку %s: %v", id, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("[INFO] Подписка %s успешно обновлена", id)
	json.NewEncoder(w).Encode(updatedSub)
}

// D - DELETE: DELETE /api/v1/subscriptions/{id}
func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/subscriptions/")

	if err := h.service.Delete(r.Context(), id); err != nil {
		log.Printf("[WARN] Не удалось удалить подписку %s (возможно, её нет в базе): %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "subscription not found"})
		return
	}

	log.Printf("[INFO] Подписка %s успешно удалена из базы данных", id)
	w.WriteHeader(http.StatusNoContent)
}

// L - LIST: GET /api/v1/subscriptions
func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	filterUserID := r.URL.Query().Get("user_id")

	list, err := h.service.List(r.Context(), filterUserID)
	if err != nil {
		log.Printf("[ERROR] Ошибка при получении списка подписок для user_id %s: %v", filterUserID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("[INFO] Успешно отдан список подписок. Найдено записей: %d", len(list))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  list,
		"total": len(list),
	})
}

// Подсчет общей стоимости
func (h *Handler) getTotalCost(w http.ResponseWriter, r *http.Request) {
	//достаем query-параметры из url
	query := r.URL.Query()

	userID := query.Get("user_id")
	serviceName := query.Get("service_name") //option
	fromStr := query.Get("from")
	toStr := query.Get("to")

	//валидация: проверка обяз.парам-ов на пустоту
	//и длины UUID польз-ля (36 симв.)
	if len(userID) != 36 || fromStr == "" || toStr == "" {
		log.Printf("[WARN] Ошибка валидации параметров в ручке /total-price")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Validation failed: 'user_id' (valid UUIDv4), 'from' and 'to' parameters are required",
		})
		return
	}

	//вызов слоя бизнес-логики (сервис), если валидация прошла успешно
	output, err := h.service.GetTotalCost(r.Context(), userID, serviceName, fromStr, toStr)
	if err != nil {
		log.Printf("[ERROR] Не удалось рассчитать стоимость для user_id %s за период %s - %s: %v", userID, fromStr, toStr, err)
		w.WriteHeader(http.StatusBadRequest) //400
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	//возвр-ем успешный ответ клиенту
	log.Printf("[INFO] Успешный расчет стоимости для user_id %s. Итог: %d руб.", userID, output.TotalCost)
	json.NewEncoder(w).Encode(output) //код ответа 200
}