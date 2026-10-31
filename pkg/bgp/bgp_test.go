package bgp

import (
	"context"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
)

func TestAnalyzer_RPKIValidationAndHijack(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	analyzer := NewAnalyzer(bus)
	analyzer.RegisterROA("198.51.100.0/24", 64500)

	var wg sync.WaitGroup
	var receivedCount int
	var mu sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wg.Add(1)
	_, err := bus.Subscribe(ctx, TopicBGPAnomaly, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 1. Valid route announcement
	validUpdate := BGPUpdate{
		Prefix:    "198.51.100.0/24",
		OriginASN: 64500,
		ASPath:    []uint32{174, 3356, 64500},
		NextHop:   "198.51.100.1",
	}

	anomaly, err := analyzer.ProcessUpdate(ctx, validUpdate)
	if err != nil {
		t.Fatalf("ProcessUpdate failed: %v", err)
	}
	if anomaly != nil {
		t.Errorf("Expected nil anomaly for valid route, got %+v", anomaly)
	}

	route, exists := analyzer.GetRoute("198.51.100.0/24")
	if !exists || route.RPKIStatus != RPKIValid {
		t.Errorf("Expected RPKIValid route, got exists=%v status=%v", exists, route.RPKIStatus)
	}

	// 2. Hijacked announcement: Rogue AS 99999 originates the prefix
	hijackUpdate := BGPUpdate{
		Prefix:    "198.51.100.0/24",
		OriginASN: 99999,
		ASPath:    []uint32{1299, 99999},
		NextHop:   "203.0.113.5",
	}

	anomaly, err = analyzer.ProcessUpdate(ctx, hijackUpdate)
	if err != nil {
		t.Fatalf("ProcessUpdate failed: %v", err)
	}
	if anomaly == nil || anomaly.Type != AnomalyPrefixHijack {
		t.Fatalf("Expected AnomalyPrefixHijack, got %+v", anomaly)
	}
	if anomaly.Severity != "CRITICAL" {
		t.Errorf("Expected CRITICAL severity, got %s", anomaly.Severity)
	}

	wg.Wait()

	mu.Lock()
	if receivedCount != 1 {
		t.Errorf("Expected 1 anomaly published to eventbus, got %d", receivedCount)
	}
	mu.Unlock()
}

func TestAnalyzer_ExcessivePrepending(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	analyzer := NewAnalyzer(bus)
	ctx := context.Background()

	// Prepending AS 64500 six times
	prependUpdate := BGPUpdate{
		Prefix:    "203.0.113.0/24",
		OriginASN: 64500,
		ASPath:    []uint32{174, 64500, 64500, 64500, 64500, 64500, 64500},
		NextHop:   "198.51.100.2",
	}

	anomaly, err := analyzer.ProcessUpdate(ctx, prependUpdate)
	if err != nil {
		t.Fatalf("ProcessUpdate failed: %v", err)
	}
	if anomaly == nil || anomaly.Type != AnomalyExcessivePrepending {
		t.Fatalf("Expected AnomalyExcessivePrepending, got %+v", anomaly)
	}
}

func TestAnalyzer_RouteFlappingAndWithdrawal(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	analyzer := NewAnalyzer(bus)
	ctx := context.Background()

	prefix := "192.0.2.0/24"

	// Trigger 5 rapid updates to trigger route flapping
	var lastAnomaly *BGPAnomaly
	for i := 0; i < 5; i++ {
		u := BGPUpdate{
			Prefix:    prefix,
			OriginASN: 65001,
			ASPath:    []uint32{2914, 65001},
			NextHop:   "192.0.2.1",
		}
		var err error
		lastAnomaly, err = analyzer.ProcessUpdate(ctx, u)
		if err != nil {
			t.Fatalf("ProcessUpdate iteration %d failed: %v", i, err)
		}
	}

	if lastAnomaly == nil || lastAnomaly.Type != AnomalyRouteFlapping {
		t.Fatalf("Expected AnomalyRouteFlapping after 5 rapid updates, got %+v", lastAnomaly)
	}

	// Test withdrawal
	withdrawal := BGPUpdate{
		Prefix:       prefix,
		IsWithdrawal: true,
	}
	_, err := analyzer.ProcessUpdate(ctx, withdrawal)
	if err != nil {
		t.Fatalf("Withdrawal failed: %v", err)
	}

	_, exists := analyzer.GetRoute(prefix)
	if exists {
		t.Errorf("Expected route to be removed after withdrawal")
	}
}
