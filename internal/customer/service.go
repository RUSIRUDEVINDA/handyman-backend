package customer

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	UpsertProfile(ctx context.Context, userID string, req UpsertCustomerRequest) (*Customer, error)
	GetProfile(ctx context.Context, userID string) (*Customer, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) UpsertProfile(ctx context.Context, userID string, req UpsertCustomerRequest) (*Customer, error) {
	now := time.Now().UTC()
	customer := &Customer{
		ID:        uuid.NewString(),
		UserID:    userID,
		FullName:  req.FullName,
		Phone:     req.Phone,
		Address:   req.Address,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Upsert(ctx, customer); err != nil {
		return nil, err
	}
	return s.repo.GetByUserID(ctx, userID)
}

func (s *service) GetProfile(ctx context.Context, userID string) (*Customer, error) {
	return s.repo.GetByUserID(ctx, userID)
}
