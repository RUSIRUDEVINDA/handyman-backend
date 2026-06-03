package events

import "time"

type EventType string

const (
	BookingCreated     EventType = "booking.created"
	BookingConfirmed   EventType = "booking.confirmed"
	BookingCompleted   EventType = "booking.completed"
	BookingCancelled   EventType = "booking.cancelled"
	PaymentIntentMade  EventType = "payment.intent_created"
	PaymentSucceeded   EventType = "payment.succeeded"
	LocationUpdated    EventType = "location.updated"
	ReviewCreated      EventType = "review.created"
	HandymanStatusMade EventType = "handyman.status_updated"
)

type Event struct {
	Type      EventType `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}
