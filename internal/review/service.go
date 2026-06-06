package review

import (
	"context"
	"errors"
	"time"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/booking"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/handyman"
	"github.com/google/uuid"
)

type Service interface {
	CreateReview(ctx context.Context, customerID string, req CreateReviewRequest) (*Review, error)
	ListForHandyman(ctx context.Context, handymanID string) ([]Review, error)
}

type service struct {
	repo         Repository
	bookingRepo  booking.Repository
	handymanRepo handyman.Repository
	bus          events.Bus
}

func NewService(repo Repository, bookingRepo booking.Repository, handymanRepo handyman.Repository, bus events.Bus) Service {
	return &service{repo: repo, bookingRepo: bookingRepo, handymanRepo: handymanRepo, bus: bus}
}

func (s *service) CreateReview(ctx context.Context, customerID string, req CreateReviewRequest) (*Review, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	bookingRecord, err := s.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		return nil, errors.New("booking not found")
	}
	if bookingRecord.CustomerID != customerID {
		return nil, errors.New("only the booking customer can review")
	}
	if bookingRecord.HandymanID == nil || *bookingRecord.HandymanID != req.HandymanID {
		return nil, errors.New("review handyman does not match booking handyman")
	}
	if bookingRecord.Status != booking.StatusCompleted {
		return nil, errors.New("booking must be completed before review")
	}
	review := &Review{
		ID:         uuid.NewString(),
		BookingID:  req.BookingID,
		CustomerID: customerID,
		HandymanID: req.HandymanID,
		Rating:     req.Rating,
		Comment:    req.Comment,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	stats, err := s.repo.StatsForHandyman(ctx, req.HandymanID)
	if err == nil {
		_ = s.handymanRepo.UpdateRating(ctx, req.HandymanID, stats.Average, stats.Count)
	}

	s.bus.Publish(events.ReviewCreated, review)
	return review, nil
}

func (s *service) ListForHandyman(ctx context.Context, handymanID string) ([]Review, error) {
	return s.repo.ListForHandyman(ctx, handymanID)
}
