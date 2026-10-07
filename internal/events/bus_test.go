package events_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"sitecraft/internal/events"
	"sitecraft/internal/models"
)

func TestEventBusPublishSubscribe(t *testing.T) {
	bus := events.NewMemoryBus()

	var called atomic.Int32
	bus.Subscribe(events.EventLeadCreated, func(ctx context.Context, e events.Event) error {
		leadEvt, ok := e.(events.LeadCreatedEvent)
		if !ok {
			t.Fatalf("expected LeadCreatedEvent, got %T", e)
		}
		if leadEvt.Lead.Name != "Иван" {
			t.Errorf("expected name Иван, got %s", leadEvt.Lead.Name)
		}
		called.Add(1)
		return nil
	})

	bus.Publish(context.Background(), events.LeadCreatedEvent{
		Lead: models.Lead{Name: "Иван"},
	})

	time.Sleep(50 * time.Millisecond)

	if called.Load() != 1 {
		t.Fatalf("expected handler to be called once, got %d", called.Load())
	}
}
