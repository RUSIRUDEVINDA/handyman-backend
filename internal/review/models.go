package review

import "time"

type Review struct {
	ID         string    `json:"id"`
	BookingID  string    `json:"booking_id"`
	CustomerID string    `json:"customer_id"`
	HandymanID string    `json:"handyman_id"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateReviewRequest struct {
	BookingID  string `json:"booking_id"`
	HandymanID string `json:"handyman_id"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
}

type RatingStats struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
