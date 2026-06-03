package payment

import "time"

type Status string

const (
	StatusRequiresPaymentMethod Status = "REQUIRES_PAYMENT_METHOD"
	StatusRequiresConfirmation  Status = "REQUIRES_CONFIRMATION"
	StatusRequiresAction        Status = "REQUIRES_ACTION"
	StatusProcessing            Status = "PROCESSING"
	StatusSucceeded             Status = "SUCCEEDED"
	StatusCanceled              Status = "CANCELED"
	StatusFailed                Status = "FAILED"
)

type Payment struct {
	ID                    string    `json:"id"`
	BookingID             string    `json:"booking_id"`
	CustomerID            string    `json:"customer_id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id"`
	AmountCents           int64     `json:"amount_cents"`
	Currency              string    `json:"currency"`
	Status                Status    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreateIntentRequest struct {
	BookingID   string `json:"booking_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type CreateIntentResponse struct {
	ID                    string  `json:"id"`
	StripePaymentIntentID string  `json:"stripe_payment_intent_id"`
	ClientSecret          string  `json:"client_secret"`
	Payment               Payment `json:"payment"`
}
