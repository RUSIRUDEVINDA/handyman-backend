package payment

import "time"

type Method string

const (
	MethodPayHere Method = "PAYHERE"
	MethodCash    Method = "CASH"
)

type Status string

const (
	StatusPending     Status = "PENDING"
	StatusSucceeded   Status = "SUCCEEDED"
	StatusCanceled    Status = "CANCELED"
	StatusFailed      Status = "FAILED"
	StatusChargedback Status = "CHARGEDBACK"
)

type Payment struct {
	ID                string    `json:"id"`
	BookingID         string    `json:"booking_id"`
	CustomerID        string    `json:"customer_id"`
	Method            Method    `json:"method"`
	Provider          string    `json:"provider"`
	ProviderPaymentID *string   `json:"provider_payment_id,omitempty"`
	ProviderOrderID   *string   `json:"provider_order_id,omitempty"`
	AmountCents       int64     `json:"amount_cents"`
	Currency          string    `json:"currency"`
	Status            Status    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreatePayHereCheckoutRequest struct {
	BookingID   string `json:"booking_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	Items       string `json:"items"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
}

type PayHereCheckoutResponse struct {
	ActionURL string            `json:"action_url"`
	Method    string            `json:"method"`
	Fields    map[string]string `json:"fields"`
	Payment   Payment           `json:"payment"`
}

type CreateCashPaymentRequest struct {
	BookingID   string `json:"booking_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type PayHereNotification struct {
	MerchantID      string
	OrderID         string
	PaymentID       string
	PayHereAmount   string
	PayHereCurrency string
	StatusCode      string
	MD5Sig          string
	Method          string
	StatusMessage   string
	CardHolderName  string
	CardNo          string
	CardExpiry      string
}
