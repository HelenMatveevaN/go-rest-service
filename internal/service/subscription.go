package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"go-rest-service/internal/domain"
)

type SubscriptionService struct {
	repo domain.SubscriptionRepository
}

func NewSubscriptionService(repo domain.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, input domain.CreateInput) (domain.Subscription, error) {
	log.Printf("[DEBUG] Сервис: Начало создания подписки на сервис '%s' для пользователя: %s", input.ServiceName, input.UserID)

	startDate, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		log.Printf("[ERROR] Сервис: Ошибка парсинга start_date '%s': %v", input.StartDate, err)
		return domain.Subscription{}, err
	}

	var endDate *time.Time
	if input.EndDate != nil {
		parsedEnd, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			log.Printf("[ERROR] Сервис: Ошибка парсинга end_date '%s': %v", *input.EndDate, err)
			return domain.Subscription{}, err
		}
		endDate = &parsedEnd
	}

	if endDate != nil && endDate.Before(startDate) {
		log.Printf("[WARN] Сервис: Нарушена бизнес-логика: end_date (%s) раньше start_date (%s)", *input.EndDate, input.StartDate)
		return domain.Subscription{}, domain.ErrInvalidDates
	}

	// Сгенерируем UUID
	realUUID := uuid.New().String()

	newSub := domain.Subscription{
		ID:          realUUID,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Save(ctx, newSub); err != nil {
		log.Printf("[ERROR] Сервис: Ошибка сохранения подписки в репозиторий: %v", err)
		return domain.Subscription{}, err
	}

	log.Printf("[DEBUG] Сервис: Подписка успешно сформирована и передана в репозиторий с ID: %s", newSub.ID)
	return newSub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	log.Printf("[DEBUG] Сервис: Запрос подписки по ID: %s", id)
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Update(ctx context.Context, id string, input domain.UpdateInput) (domain.Subscription, error) {
	log.Printf("[DEBUG] Сервис: Начало процесса обновления подписки с ID: %s", id)

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[WARN] Сервис: Не удалось обновить, подписка %s не найдена: %v", id, err)
		return domain.Subscription{}, err
	}

	if input.ServiceName != nil {
		sub.ServiceName = *input.ServiceName
	}
	if input.Price != nil {
		sub.Price = *input.Price
	}
	if input.EndDate != nil {
		if *input.EndDate == nil {
			sub.EndDate = nil
			log.Printf("[DEBUG] Сервис: Сброс даты окончания подписки %s (подписка стала бессрочной)", id)
		} else {
			parsedEnd, err := time.Parse("01-2006", **input.EndDate)
			if err != nil {
				log.Printf("[ERROR] Сервис: Ошибка парсинга новой end_date '%s': %v", **input.EndDate, err)
				return domain.Subscription{}, err
			}
			if parsedEnd.Before(sub.StartDate) {
				log.Printf("[WARN] Сервис: Нарушена бизнес-логика при обновлении: новая end_date раньше start_date")
				return domain.Subscription{}, domain.ErrInvalidDates
			}
			sub.EndDate = &parsedEnd
		}
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		log.Printf("[ERROR] Сервис: Не удалось обновить запись %s в репозитории: %v", id, err)
		return domain.Subscription{}, err
	}

	log.Printf("[DEBUG] Сервис: Запись %s успешно обновлена", id)
	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	log.Printf("[DEBUG] Сервис: Удаление подписки по ID: %s", id)
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, userID string) ([]domain.Subscription, error) {
	log.Printf("[DEBUG] Сервис: Запрос списка подписок для user_id: %s", userID)
	return s.repo.List(ctx, userID)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID, serviceName, fromStr, toStr string) (domain.GetTotalCostOutput, error) {
	log.Printf("[DEBUG] Сервис: Начало расчета стоимости для user_id: %s за период %s - %s (фильтр сервиса: '%s')", userID, fromStr, toStr, serviceName)

	// 1. Валидация дат
	fromTime, err := time.Parse("01-2006", fromStr)
	if err != nil {
		return domain.GetTotalCostOutput{}, fmt.Errorf("invalid 'from' date format: %w", err)
	}

	toTime, err := time.Parse("01-2006", toStr)
	if err != nil {
		return domain.GetTotalCostOutput{}, fmt.Errorf("invalid 'to' date format: %w", err)
	}

	// Валидация: begin не м.б. > end
	if toTime.Before(fromTime) {
		return domain.GetTotalCostOutput{}, fmt.Errorf("'from' date cannot be after 'to' date")
	}

	// 2. Достаем подписки из репозитория
	subs, err := s.repo.GetByFilters(ctx, userID, serviceName)
	if err != nil {
		return domain.GetTotalCostOutput{}, err
	}

	log.Printf("[DEBUG] Сервис: Найдено %d подписок для расчета стоимости", len(subs))

	totalCost := 0

	// 3. Подсчет стоимости подписки
	for _, sub := range subs {
		//реальная дата начала пересечения
		//get max date between началом подписки и началом выбранного периода
		startIntersection := sub.StartDate
		if fromTime.After(startIntersection) {
			startIntersection = fromTime
		}

		//реальная дата окончания пересечения
		//get min date between окончанием подписки и окончанием выбранного периода
		endIntersection := toTime
		if sub.EndDate != nil && sub.EndDate.Before(endIntersection) {
			endIntersection = *sub.EndDate
		}

		//Если endIntersection < startIntersection, пересечений нет
		if endIntersection.Before(startIntersection){
			continue
		}

		//Считаем кол-во месяцев м.у. startIntersection и endIntersection
		//включая пограничные месяцы
		yearsDiff := endIntersection.Year() - startIntersection.Year()
		monthsDiff := int(endIntersection.Month()) - int(startIntersection.Month())

		monthsCount := yearsDiff*12 + monthsDiff + 1

		//Прибавляем к общей сумме
		totalCost += monthsCount * sub.Price
	}

	log.Printf("[DEBUG] Сервис: Расчет завершен. Суммарная стоимость: %d руб.", totalCost)
	return domain.GetTotalCostOutput{TotalCost: totalCost}, nil
}