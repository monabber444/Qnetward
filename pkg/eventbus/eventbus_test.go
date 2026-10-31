package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryBusPubSub(t *testing.T) {
	bus := NewMemoryBus()
	defer bus.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	var received Message
	sub, err := bus.Subscribe(ctx, TopicTelemetryProbe, func(c context.Context, msg Message) error {
		received = msg
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	payload := []byte(`{"probe_id":"test-123","rtt_ms":12.5}`)
	if err := bus.Publish(ctx, TopicTelemetryProbe, payload); err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	wg.Wait()

	if string(received.Data) != string(payload) {
		t.Fatalf("expected payload %s, got %s", payload, received.Data)
	}
	if received.Subject != TopicTelemetryProbe {
		t.Fatalf("expected subject %s, got %s", TopicTelemetryProbe, received.Subject)
	}

	if err := sub.Unsubscribe(); err != nil {
		t.Fatalf("failed to unsubscribe: %v", err)
	}
}

func TestMemoryBusWildcards(t *testing.T) {
	bus := NewMemoryBus()
	defer bus.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(2)

	var count int
	var mu sync.Mutex

	_, err := bus.Subscribe(ctx, "telemetry.*", func(c context.Context, msg Message) error {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	_ = bus.Publish(ctx, "telemetry.probe", []byte("1"))
	_ = bus.Publish(ctx, "telemetry.mesh", []byte("2"))
	_ = bus.Publish(ctx, "agent.heartbeat", []byte("3")) // Should not match

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for wildcard events")
	}

	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Fatalf("expected 2 received messages, got %d", count)
	}
}

func TestMemoryBusClosed(t *testing.T) {
	bus := NewMemoryBus()
	_ = bus.Close()

	ctx := context.Background()
	if err := bus.Publish(ctx, "test", []byte("data")); err != ErrClosed {
		t.Fatalf("expected ErrClosed on publish, got %v", err)
	}
	if _, err := bus.Subscribe(ctx, "test", func(c context.Context, msg Message) error { return nil }); err != ErrClosed {
		t.Fatalf("expected ErrClosed on subscribe, got %v", err)
	}
}
