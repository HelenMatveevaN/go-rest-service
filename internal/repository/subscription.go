package repository

import (
	"context"
	"fmt"
	"sync"
	"go-rest-service/internal/domain"
)

type MemoryRepository struct {
	mu				sync.RWMutex
	subscriptions	map[string]domain.Subscription
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		subscriptions: make(map[string]domain.Subscription),
	}
}

func (r *MemoryRepository) Save(ctx context.Context, sub domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subscriptions[sub.ID] = sub
	return nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sub, exists := r.subscriptions[id]
	if !exists {
		return domain.Subscription{}, fmt.Errorf("subscription not found")
	}
	return sub, nil
}

func (r *MemoryRepository) Update(ctx context.Context, sub domain.Subscription) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, exists := r.subscriptions[sub.ID]; !exists {
		return fmt.Errorf("subscription not found")
	}
	r.subscriptions[sub.ID] = sub
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.subscriptions[id]; !exists {
		return fmt.Errorf("subscription not found")
	}
	delete(r.subscriptions, id)
	return nil
}

func (r *MemoryRepository) List(ctx context.Context, userID string) ([]domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Subscription, 0)
	for _, sub := range r.subscriptions {
		if userID != "" && sub.UserID != userID {
			continue
		}
		list = append(list, sub)
	}
	return list, nil
}

func (r *MemoryRepository) GetByFilters(ctx context.Context, userID, ServiceName string) ([]domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Subscription, 0)
	for _, sub := range r.subscriptions {
		if sub.UserID != userID {
			continue
		}

		if ServiceName != "" && sub.ServiceName != ServiceName {
			continue
		}

		list = append(list, sub)
	}

	return list, nil
}