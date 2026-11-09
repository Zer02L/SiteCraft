package events

import (
	"context"
	"log/slog"
	"sync"
	"time"
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
	wg       sync.WaitGroup
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
// The context is detached from request cancellation using context.WithoutCancel,
// given a 15-second execution budget, and tracked via WaitGroup for graceful shutdown.
func (b *MemoryBus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[event.EventName()]...)
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	asyncCtx := context.WithoutCancel(ctx)

	for _, h := range handlers {
		handler := h
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("EventBus Panic", "event", event.EventName(), "panic", r)
				}
			}()

			timeoutCtx, cancel := context.WithTimeout(asyncCtx, 15*time.Second)
			defer cancel()

			if err := handler(timeoutCtx, event); err != nil {
				slog.Error("EventBus handler error", "event", event.EventName(), "error", err)
			}
		}()
	}
}

// Shutdown waits for active event handlers to complete or until the context expires.
func (b *MemoryBus) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

