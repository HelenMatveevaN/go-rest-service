package service

import (
	"context"
	"time"
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

	mockID := "sub-" + time.Now().Format("20060102150405")

	newSub := domain.Subscription{
		ID:          mockID,
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