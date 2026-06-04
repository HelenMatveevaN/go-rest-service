package service

import (
	"context"
	"fmt"
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
	startDate, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	var endDate *time.Time
	if input.EndDate != nil {
		parsedEnd, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			return domain.Subscription{}, err
		}
		endDate = &parsedEnd
	}

	if endDate != nil && endDate.Before(startDate) {
		return domain.Subscription{}, domain.ErrInvalidDates
	}

	// Вместо mockID со строкой "sub-..." генерируем настоящий UUID v4, 
	// который полностью удовлетворяет требованиям PostgreSQL к типу данных UUID.
	realUUID := uuid.New().String()
	//mockID := "sub-" + time.Now().Format("20060102150405")

	newSub := domain.Subscription{
		ID:          realUUID, // Используем валидный UUID //mockID,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Save(ctx, newSub); err != nil {
		return domain.Subscription{}, err
	}

	return newSub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Update(ctx context.Context, id string, input domain.UpdateInput) (domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
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
		} else {
			parsedEnd, err := time.Parse("01-2006", **input.EndDate)
			if err != nil {
				return domain.Subscription{}, err
			}
			if parsedEnd.Before(sub.StartDate) {
				return domain.Subscription{}, domain.ErrInvalidDates
			}
			sub.EndDate = &parsedEnd
		}
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return domain.Subscription{}, err
	}

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, userID string) ([]domain.Subscription, error) {
	return s.repo.List(ctx, userID)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID, ServiceName, fromStr, toStr string) (domain.GetTotalCostOutput, error) {
	
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
	subs, err := s.repo.GetByFilters(ctx, userID, ServiceName)
	if err != nil {
		return domain.GetTotalCostOutput{}, err
	}

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

	return domain.GetTotalCostOutput{TotalCost: totalCost}, nil
}