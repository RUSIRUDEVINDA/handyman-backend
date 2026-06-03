package booking

import (
	"context"
	"errors"
	"time"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
	"github.com/google/uuid"
)

type Service interface {
	CreateBooking(ctx context.Context, customerID string, req CreateBookingRequest) (*Booking, error)
	GetBooking(ctx context.Context, id string) (*Booking, error)
	ListCustomerBookings(ctx context.Context, customerID string) ([]Booking, error)
	ListHandymanBookings(ctx context.Context, handymanID string) ([]Booking, error)
	AssignHandyman(ctx context.Context, bookingID string, handymanID string) (*Booking, error)
	Confirm(ctx context.Context, bookingID string) (*Booking, error)
	Complete(ctx context.Context, bookingID string) (*Booking, error)
	Cancel(ctx context.Context, bookingID string) (*Booking, error)
}

type service struct {
	repo Repository
	bus  events.Bus
}

func NewService(repo Repository, bus events.Bus) Service {
	return &service{repo: repo, bus: bus}
}

func (s *service) CreateBooking(ctx context.Context, customerID string, req CreateBookingRequest) (*Booking, error) {
	if req.ServiceName == "" {
		return nil, errors.New("service_name is required")
	}
	now := time.Now().UTC()
	booking := &Booking{
		ID:          uuid.NewString(),
		CustomerID:  customerID,
		HandymanID:  req.HandymanID,
		ServiceName: req.ServiceName,
		Notes:       req.Notes,
		Status:      StatusPending,
		AmountCents: req.AmountCents,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, booking); err != nil {
		return nil, err
	}
	s.bus.Publish(events.BookingCreated, booking)
	return booking, nil
}

func (s *service) GetBooking(ctx context.Context, id string) (*Booking, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListCustomerBookings(ctx context.Context, customerID string) ([]Booking, error) {
	return s.repo.ListForCustomer(ctx, customerID)
}

func (s *service) ListHandymanBookings(ctx context.Context, handymanID string) ([]Booking, error) {
	return s.repo.ListForHandyman(ctx, handymanID)
}

func (s *service) AssignHandyman(ctx context.Context, bookingID string, handymanID string) (*Booking, error) {
	if err := s.repo.AssignHandyman(ctx, bookingID, handymanID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, bookingID)
}

func (s *service) Confirm(ctx context.Context, bookingID string) (*Booking, error) {
	return s.transition(ctx, bookingID, StatusConfirmed, events.BookingConfirmed)
}

func (s *service) Complete(ctx context.Context, bookingID string) (*Booking, error) {
	return s.transition(ctx, bookingID, StatusCompleted, events.BookingCompleted)
}

func (s *service) Cancel(ctx context.Context, bookingID string) (*Booking, error) {
	return s.transition(ctx, bookingID, StatusCancelled, events.BookingCancelled)
}

func (s *service) transition(ctx context.Context, bookingID string, status Status, eventType events.EventType) (*Booking, error) {
	if err := s.repo.UpdateStatus(ctx, bookingID, status); err != nil {
		return nil, err
	}
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	s.bus.Publish(eventType, booking)
	return booking, nil
}
