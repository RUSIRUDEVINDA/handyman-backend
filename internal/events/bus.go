package events

import (
	"sync"
	"time"
)

type Bus interface {
	Publish(eventType EventType, payload any)
	Subscribe(eventType EventType) <-chan Event
	Close()
}

type memoryBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]chan Event
	closed      bool
}

func NewBus() Bus {
	return &memoryBus{subscribers: make(map[EventType][]chan Event)}
}

func (b *memoryBus) Publish(eventType EventType, payload any) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	event := Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	for _, subscriber := range b.subscribers[eventType] {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (b *memoryBus) Subscribe(eventType EventType) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, 100)
	if b.closed {
		close(ch)
		return ch
	}

	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
	return ch
}

func (b *memoryBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	for _, subscribers := range b.subscribers {
		for _, subscriber := range subscribers {
			close(subscriber)
		}
	}
	b.subscribers = make(map[EventType][]chan Event)
	b.closed = true
}
