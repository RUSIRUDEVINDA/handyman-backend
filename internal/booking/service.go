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
	Approve(ctx context.Context, bookingID string, handymanID string) (*Booking, error)
	Reject(ctx context.Context, bookingID string, handymanID string) (*Booking, error)
	Confirm(ctx context.Context, bookingID string) (*Booking, error)
	MarkPaymentPending(ctx context.Context, bookingID string) (*Booking, error)
	Start(ctx context.Context, bookingID string, handymanID string) (*Booking, error)
	Complete(ctx context.Context, bookingID string, handymanID string) (*Booking, error)
	Cancel(ctx context.Context, bookingID string, userID string) (*Booking, error)
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
		Status:      StatusRequested,
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

func (s *service) Approve(ctx context.Context, bookingID string, handymanID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status != StatusRequested {
		return nil, errors.New("only requested bookings can be approved")
	}
	if booking.HandymanID != nil && *booking.HandymanID != handymanID {
		return nil, errors.New("booking is assigned to another handyman")
	}
	if booking.HandymanID == nil {
		if err := s.repo.AssignHandyman(ctx, bookingID, handymanID); err != nil {
			return nil, err
		}
	}
	return s.transition(ctx, bookingID, StatusApproved, events.BookingApproved)
}

func (s *service) Reject(ctx context.Context, bookingID string, handymanID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status != StatusRequested {
		return nil, errors.New("only requested bookings can be rejected")
	}
	if booking.HandymanID != nil && *booking.HandymanID != handymanID {
		return nil, errors.New("booking is assigned to another handyman")
	}
	return s.transition(ctx, bookingID, StatusRejected, events.BookingRejected)
}

func (s *service) Confirm(ctx context.Context, bookingID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status != StatusPaymentPending && booking.Status != StatusApproved {
		return nil, errors.New("only approved or payment pending bookings can be confirmed")
	}
	return s.transition(ctx, bookingID, StatusConfirmed, events.BookingConfirmed)
}

func (s *service) MarkPaymentPending(ctx context.Context, bookingID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status != StatusApproved {
		return nil, errors.New("only approved bookings can move to payment pending")
	}
	return s.transition(ctx, bookingID, StatusPaymentPending, events.PaymentCreated)
}

func (s *service) Start(ctx context.Context, bookingID string, handymanID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status != StatusConfirmed {
		return nil, errors.New("only confirmed bookings can be started")
	}
	if booking.HandymanID == nil || *booking.HandymanID != handymanID {
		return nil, errors.New("only the assigned handyman can start this booking")
	}
	return s.transition(ctx, bookingID, StatusInProgress, events.BookingStarted)
}

func (s *service) Complete(ctx context.Context, bookingID string, handymanID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status != StatusInProgress {
		return nil, errors.New("only in-progress bookings can be completed")
	}
	if booking.HandymanID == nil || *booking.HandymanID != handymanID {
		return nil, errors.New("only the assigned handyman can complete this booking")
	}
	return s.transition(ctx, bookingID, StatusCompleted, events.BookingCompleted)
}

func (s *service) Cancel(ctx context.Context, bookingID string, userID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status == StatusCompleted || booking.Status == StatusCancelled {
		return nil, errors.New("completed or cancelled bookings cannot be cancelled")
	}
	if booking.CustomerID != userID && (booking.HandymanID == nil || *booking.HandymanID != userID) {
		return nil, errors.New("only the customer or assigned handyman can cancel this booking")
	}
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
