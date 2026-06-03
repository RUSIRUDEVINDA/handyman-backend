package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
	fiberws "github.com/gofiber/contrib/websocket"
)

type Hub struct {
	bus     events.Bus
	mu      sync.RWMutex
	clients map[*fiberws.Conn]struct{}
}

func NewHub(bus events.Bus) *Hub {
	return &Hub{
		bus:     bus,
		clients: make(map[*fiberws.Conn]struct{}),
	}
}

func (h *Hub) Add(conn *fiberws.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Remove(conn *fiberws.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(message any) {
	payload, err := json.Marshal(message)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if err := client.WriteMessage(fiberws.TextMessage, payload); err != nil {
			log.Printf("websocket write failed: %v", err)
		}
	}
}

func (h *Hub) Start(ctx context.Context) {
	eventTypes := []events.EventType{
		events.BookingCreated,
		events.BookingConfirmed,
		events.BookingCompleted,
		events.BookingCancelled,
		events.PaymentSucceeded,
		events.LocationUpdated,
		events.ReviewCreated,
	}

	for _, eventType := range eventTypes {
		ch := h.bus.Subscribe(eventType)
		go func(events <-chan events.Event) {
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-events:
					if !ok {
						return
					}
					h.Broadcast(event)
				}
			}
		}(ch)
	}
}
