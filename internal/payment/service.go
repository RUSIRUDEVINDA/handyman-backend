package payment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type Service interface {
	CreatePaymentIntent(ctx context.Context, customerID string, req CreateIntentRequest) (*CreateIntentResponse, error)
	GetByBookingID(ctx context.Context, bookingID string) (*Payment, error)
	MarkPaymentSucceeded(ctx context.Context, paymentIntentID string, amount int64) error
}

type service struct {
	repo Repository
	bus  events.Bus
}

func NewService(secretKey string, repo Repository, bus events.Bus) Service {
	if secretKey != "" {
		stripe.Key = secretKey
	}
	return &service{repo: repo, bus: bus}
}

func (s *service) CreatePaymentIntent(ctx context.Context, customerID string, req CreateIntentRequest) (*CreateIntentResponse, error) {
	if req.BookingID == "" {
		return nil, errors.New("booking_id is required")
	}
	if req.AmountCents <= 0 {
		return nil, errors.New("amount_cents must be greater than zero")
	}
	if req.Currency == "" {
		req.Currency = "usd"
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.AmountCents),
		Currency: stripe.String(req.Currency),
		Metadata: map[string]string{
			"booking_id": req.BookingID,
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	payment := Payment{
		ID:                    uuid.NewString(),
		BookingID:             req.BookingID,
		CustomerID:            customerID,
		StripePaymentIntentID: intent.ID,
		AmountCents:           intent.Amount,
		Currency:              strings.ToLower(string(intent.Currency)),
		Status:                normalizeStripeStatus(string(intent.Status)),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := s.repo.Create(ctx, &payment); err != nil {
		return nil, err
	}

	response := &CreateIntentResponse{
		ID:                    payment.ID,
		StripePaymentIntentID: intent.ID,
		ClientSecret:          intent.ClientSecret,
		Payment:               payment,
	}
	s.bus.Publish(events.PaymentIntentMade, response)
	return response, nil
}

func (s *service) GetByBookingID(ctx context.Context, bookingID string) (*Payment, error) {
	return s.repo.GetByBookingID(ctx, bookingID)
}

func (s *service) MarkPaymentSucceeded(ctx context.Context, paymentIntentID string, amount int64) error {
	if err := s.repo.UpdateStatusByIntentID(ctx, paymentIntentID, StatusSucceeded); err != nil {
		return err
	}
	s.bus.Publish(events.PaymentSucceeded, fiberSafePaymentPayload{
		PaymentIntentID: paymentIntentID,
		Amount:          amount,
	})
	return nil
}

func normalizeStripeStatus(status string) Status {
	normalized := strings.ToUpper(status)
	switch normalized {
	case string(StatusRequiresPaymentMethod),
		string(StatusRequiresConfirmation),
		string(StatusRequiresAction),
		string(StatusProcessing),
		string(StatusSucceeded),
		string(StatusCanceled):
		return Status(normalized)
	default:
		return StatusFailed
	}
}

type fiberSafePaymentPayload struct {
	PaymentIntentID string `json:"payment_intent_id"`
	Amount          int64  `json:"amount"`
}
