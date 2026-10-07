package events

import (
	"context"
	"log"
	"sync"
)

// Event represents any domain event within the application.
type Event interface {
	EventName() string
}

// Handler defines a callback for processing an event asynchronously or synchronously.
type Handler func(ctx context.Context, event Event) error

// Bus defines the event publisher and subscriber interface.
type Bus interface {
	Subscribe(eventName string, handler Handler)
	Publish(ctx context.Context, event Event)
}

// MemoryBus is a thread-safe in-memory Event Bus.
type MemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

var (
	defaultBus *MemoryBus
	once       sync.Once
)

// Default returns the singleton in-memory event bus.
func Default() *MemoryBus {
	once.Do(func() {
		defaultBus = NewMemoryBus()
	})
	return defaultBus
}

// NewMemoryBus creates a new in-memory event bus.
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers an event handler for a specific event name.
func (b *MemoryBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish distributes the event asynchronously to all subscribed handlers.
func (b *MemoryBus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[event.EventName()]...)
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	for _, h := range handlers {
		handler := h
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[EventBus Panic] Event: %s, Error: %v\n", event.EventName(), r)
				}
			}()
			if err := handler(ctx, event); err != nil {
				log.Printf("[EventBus Error] Event %s failed: %v\n", event.EventName(), err)
			}
		}()
	}
}
