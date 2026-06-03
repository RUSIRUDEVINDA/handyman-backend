package handyman

import (
	"context"
	"time"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
	"github.com/google/uuid"
)

type Service interface {
	UpsertProfile(ctx context.Context, userID string, req UpsertHandymanRequest) (*Handyman, error)
	GetProfile(ctx context.Context, userID string) (*Handyman, error)
	Search(ctx context.Context, skill string) ([]Handyman, error)
}

type service struct {
	repo Repository
	bus  events.Bus
}

func NewService(repo Repository, bus events.Bus) Service {
	return &service{repo: repo, bus: bus}
}

func (s *service) UpsertProfile(ctx context.Context, userID string, req UpsertHandymanRequest) (*Handyman, error) {
	now := time.Now().UTC()
	if req.Status == "" {
		req.Status = StatusOffline
	}

	profile := &Handyman{
		ID:          uuid.NewString(),
		UserID:      userID,
		FullName:    req.FullName,
		Phone:       req.Phone,
		Bio:         req.Bio,
		Skills:      req.Skills,
		HourlyRate:  req.HourlyRate,
		Status:      req.Status,
		Rating:      0,
		ReviewCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Upsert(ctx, profile); err != nil {
		return nil, err
	}

	s.bus.Publish(events.HandymanStatusMade, fiberSafeStatusPayload{UserID: userID, Status: string(req.Status)})
	return s.repo.GetByUserID(ctx, userID)
}

func (s *service) GetProfile(ctx context.Context, userID string) (*Handyman, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *service) Search(ctx context.Context, skill string) ([]Handyman, error) {
	return s.repo.Search(ctx, skill)
}

type fiberSafeStatusPayload struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}
