package booking

import "time"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusCompleted Status = "COMPLETED"
	StatusCancelled Status = "CANCELLED"
)

type Booking struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	HandymanID  *string   `json:"handyman_id,omitempty"`
	ServiceName string    `json:"service_name"`
	Notes       string    `json:"notes"`
	Status      Status    `json:"status"`
	AmountCents int64     `json:"amount_cents"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateBookingRequest struct {
	HandymanID  *string `json:"handyman_id"`
	ServiceName string  `json:"service_name"`
	Notes       string  `json:"notes"`
	AmountCents int64   `json:"amount_cents"`
}

type AssignHandymanRequest struct {
	HandymanID string `json:"handyman_id"`
}
