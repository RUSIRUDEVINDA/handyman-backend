package notification

import (
	"context"
	"fmt"
	"log"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
)

type Worker struct {
	bus     events.Bus
	service Service
}

func NewWorker(bus events.Bus, service Service) *Worker {
	return &Worker{bus: bus, service: service}
}

func (w *Worker) Start(ctx context.Context) {
	bookingCreated := w.bus.Subscribe(events.BookingCreated)
	bookingApproved := w.bus.Subscribe(events.BookingApproved)
	bookingConfirmed := w.bus.Subscribe(events.BookingConfirmed)
	bookingStarted := w.bus.Subscribe(events.BookingStarted)
	paymentSucceeded := w.bus.Subscribe(events.PaymentSucceeded)

	log.Println("notification worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("notification worker stopped")
			return
		case event, ok := <-bookingCreated:
			if ok {
				go w.notify("New job request", event)
			}
		case event, ok := <-bookingApproved:
			if ok {
				go w.notify("Booking approved - payment can continue", event)
			}
		case event, ok := <-bookingConfirmed:
			if ok {
				go w.notify("Booking confirmed", event)
			}
		case event, ok := <-bookingStarted:
			if ok {
				go w.notify("Job started", event)
			}
		case event, ok := <-paymentSucceeded:
			if ok {
				go w.notify("Payment succeeded", event)
			}
		}
	}
}

func (w *Worker) notify(subject string, event events.Event) {
	body := fmt.Sprintf("%s at %s", event.Type, event.Timestamp.Format("2006-01-02 15:04:05"))
	if err := w.service.SendEmail("ops@handyman.local", subject, body); err != nil {
		log.Printf("notification failed: %v", err)
	}
}
