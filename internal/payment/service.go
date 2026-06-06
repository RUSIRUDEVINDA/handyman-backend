package payment

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/booking"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
	"github.com/google/uuid"
)

type PayHereConfig struct {
	MerchantID     string
	MerchantSecret string
	CheckoutURL    string
	ReturnURL      string
	CancelURL      string
	NotifyURL      string
}

type Service interface {
	CreatePayHereCheckout(ctx context.Context, customerID string, req CreatePayHereCheckoutRequest) (*PayHereCheckoutResponse, error)
	CreateCashPayment(ctx context.Context, customerID string, req CreateCashPaymentRequest) (*Payment, error)
	GetByBookingID(ctx context.Context, bookingID string) (*Payment, error)
	MarkCashCollected(ctx context.Context, paymentID string, handymanID string) error
	HandlePayHereNotification(ctx context.Context, notification PayHereNotification) error
}

type service struct {
	repo          Repository
	bookingRepo   booking.Repository
	bus           events.Bus
	payHereConfig PayHereConfig
}

func NewService(repo Repository, bookingRepo booking.Repository, bus events.Bus, payHereConfig PayHereConfig) Service {
	return &service{repo: repo, bookingRepo: bookingRepo, bus: bus, payHereConfig: payHereConfig}
}

func (s *service) CreatePayHereCheckout(ctx context.Context, customerID string, req CreatePayHereCheckoutRequest) (*PayHereCheckoutResponse, error) {
	if s.payHereConfig.MerchantID == "" || s.payHereConfig.MerchantSecret == "" {
		return nil, errors.New("PayHere merchant id and merchant secret are required")
	}
	if req.BookingID == "" {
		return nil, errors.New("booking_id is required")
	}
	bookingRecord, err := s.validatePayableBooking(ctx, customerID, req.BookingID)
	if err != nil {
		return nil, err
	}
	if req.AmountCents == 0 {
		req.AmountCents = bookingRecord.AmountCents
	}
	if req.AmountCents <= 0 {
		return nil, errors.New("amount_cents must be greater than zero")
	}
	if bookingRecord.AmountCents > 0 && req.AmountCents != bookingRecord.AmountCents {
		return nil, errors.New("payment amount must match booking amount")
	}
	if req.Currency == "" {
		req.Currency = "LKR"
	}
	if req.Items == "" {
		req.Items = "Handyman service booking"
	}
	if req.Country == "" {
		req.Country = "Sri Lanka"
	}

	now := time.Now().UTC()
	paymentID := uuid.NewString()
	providerOrderID := paymentID
	payment := Payment{
		ID:              paymentID,
		BookingID:       req.BookingID,
		CustomerID:      customerID,
		Method:          MethodPayHere,
		Provider:        "PAYHERE",
		ProviderOrderID: &providerOrderID,
		AmountCents:     req.AmountCents,
		Currency:        strings.ToUpper(req.Currency),
		Status:          StatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, &payment); err != nil {
		return nil, err
	}
	if err := s.bookingRepo.UpdateStatus(ctx, req.BookingID, booking.StatusPaymentPending); err != nil {
		return nil, err
	}

	amount := formatAmount(req.AmountCents)
	fields := map[string]string{
		"merchant_id": s.payHereConfig.MerchantID,
		"return_url":  s.payHereConfig.ReturnURL,
		"cancel_url":  s.payHereConfig.CancelURL,
		"notify_url":  s.payHereConfig.NotifyURL,
		"order_id":    providerOrderID,
		"items":       req.Items,
		"currency":    payment.Currency,
		"amount":      amount,
		"first_name":  req.FirstName,
		"last_name":   req.LastName,
		"email":       req.Email,
		"phone":       req.Phone,
		"address":     req.Address,
		"city":        req.City,
		"country":     req.Country,
		"custom_1":    req.BookingID,
		"custom_2":    customerID,
	}
	fields["hash"] = s.generateCheckoutHash(providerOrderID, amount, payment.Currency)

	response := &PayHereCheckoutResponse{
		ActionURL: s.payHereConfig.CheckoutURL,
		Method:    "POST",
		Fields:    fields,
		Payment:   payment,
	}
	s.bus.Publish(events.PaymentCreated, response)
	return response, nil
}

func (s *service) CreateCashPayment(ctx context.Context, customerID string, req CreateCashPaymentRequest) (*Payment, error) {
	if req.BookingID == "" {
		return nil, errors.New("booking_id is required")
	}
	bookingRecord, err := s.validatePayableBooking(ctx, customerID, req.BookingID)
	if err != nil {
		return nil, err
	}
	if req.AmountCents == 0 {
		req.AmountCents = bookingRecord.AmountCents
	}
	if req.AmountCents <= 0 {
		return nil, errors.New("amount_cents must be greater than zero")
	}
	if bookingRecord.AmountCents > 0 && req.AmountCents != bookingRecord.AmountCents {
		return nil, errors.New("payment amount must match booking amount")
	}
	if req.Currency == "" {
		req.Currency = "LKR"
	}

	now := time.Now().UTC()
	payment := &Payment{
		ID:          uuid.NewString(),
		BookingID:   req.BookingID,
		CustomerID:  customerID,
		Method:      MethodCash,
		Provider:    "CASH",
		AmountCents: req.AmountCents,
		Currency:    strings.ToUpper(req.Currency),
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}
	if err := s.bookingRepo.UpdateStatus(ctx, req.BookingID, booking.StatusConfirmed); err != nil {
		return nil, err
	}
	if booking, err := s.bookingRepo.GetByID(ctx, req.BookingID); err == nil {
		s.bus.Publish(events.BookingConfirmed, booking)
	}
	s.bus.Publish(events.PaymentCreated, payment)
	return payment, nil
}

func (s *service) GetByBookingID(ctx context.Context, bookingID string) (*Payment, error) {
	return s.repo.GetByBookingID(ctx, bookingID)
}

func (s *service) MarkCashCollected(ctx context.Context, paymentID string, handymanID string) error {
	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return errors.New("payment not found")
	}
	if payment.Method != MethodCash {
		return errors.New("only cash payments can be marked collected manually")
	}
	bookingRecord, err := s.bookingRepo.GetByID(ctx, payment.BookingID)
	if err != nil {
		return errors.New("booking not found")
	}
	if bookingRecord.HandymanID == nil || *bookingRecord.HandymanID != handymanID {
		return errors.New("only the assigned handyman can mark cash collected")
	}
	if err := s.repo.UpdateStatus(ctx, paymentID, StatusSucceeded); err != nil {
		return err
	}
	s.bus.Publish(events.PaymentSucceeded, fiberSafePaymentPayload{
		PaymentID: paymentID,
		Method:    string(MethodCash),
	})
	return nil
}

func (s *service) HandlePayHereNotification(ctx context.Context, notification PayHereNotification) error {
	if notification.MerchantID != s.payHereConfig.MerchantID {
		return errors.New("invalid PayHere merchant_id")
	}
	if !s.verifyNotification(notification) {
		return errors.New("invalid PayHere md5 signature")
	}

	status := payHereStatus(notification.StatusCode)
	providerPaymentID := stringPtrIfNotEmpty(notification.PaymentID)
	if err := s.repo.UpdateProviderResult(ctx, notification.OrderID, providerPaymentID, status); err != nil {
		return err
	}
	payment, err := s.repo.GetByProviderOrderID(ctx, notification.OrderID)
	if err != nil {
		return err
	}
	if status == StatusSucceeded {
		if err := s.bookingRepo.UpdateStatus(ctx, payment.BookingID, booking.StatusConfirmed); err != nil {
			return err
		}
		if booking, err := s.bookingRepo.GetByID(ctx, payment.BookingID); err == nil {
			s.bus.Publish(events.BookingConfirmed, booking)
		}
	}

	eventType := events.PaymentFailed
	if status == StatusSucceeded {
		eventType = events.PaymentSucceeded
	}
	s.bus.Publish(eventType, fiberSafePaymentPayload{
		PaymentID:         notification.OrderID,
		ProviderPaymentID: notification.PaymentID,
		Method:            string(MethodPayHere),
		Status:            string(status),
	})
	return nil
}

func (s *service) validatePayableBooking(ctx context.Context, customerID string, bookingID string) (*booking.Booking, error) {
	bookingRecord, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, errors.New("booking not found")
	}
	if bookingRecord.CustomerID != customerID {
		return nil, errors.New("booking does not belong to this customer")
	}
	if bookingRecord.Status != booking.StatusApproved && bookingRecord.Status != booking.StatusPaymentPending {
		return nil, errors.New("handyman must approve the booking before payment")
	}
	return bookingRecord, nil
}

func (s *service) generateCheckoutHash(orderID string, amount string, currency string) string {
	return md5Upper(s.payHereConfig.MerchantID + orderID + amount + currency + md5Upper(s.payHereConfig.MerchantSecret))
}

func (s *service) verifyNotification(notification PayHereNotification) bool {
	localSig := md5Upper(
		notification.MerchantID +
			notification.OrderID +
			notification.PayHereAmount +
			notification.PayHereCurrency +
			notification.StatusCode +
			md5Upper(s.payHereConfig.MerchantSecret),
	)
	return strings.EqualFold(localSig, notification.MD5Sig)
}

func payHereStatus(statusCode string) Status {
	switch statusCode {
	case "2":
		return StatusSucceeded
	case "0":
		return StatusPending
	case "-1":
		return StatusCanceled
	case "-3":
		return StatusChargedback
	default:
		return StatusFailed
	}
}

func formatAmount(amountCents int64) string {
	return fmt.Sprintf("%.2f", float64(amountCents)/100)
}

func md5Upper(value string) string {
	hash := md5.Sum([]byte(value))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func stringPtrIfNotEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

type fiberSafePaymentPayload struct {
	PaymentID         string `json:"payment_id"`
	ProviderPaymentID string `json:"provider_payment_id,omitempty"`
	Method            string `json:"method"`
	Status            string `json:"status,omitempty"`
}
