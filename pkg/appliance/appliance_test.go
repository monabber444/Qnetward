package appliance

import (
	"context"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
)

func TestAppliance_ZTPFlow(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	mgr := NewManager(bus)

	// 1. Generate ZTP enrollment token
	token, err := mgr.GenerateZTPToken("tenant-retail", "probe-london-01", 10*time.Minute)
	if err != nil {
		t.Fatalf("GenerateZTPToken failed: %v", err)
	}
	if len(token.Token) < 16 {
		t.Errorf("Expected robust token length, got %s", token.Token)
	}

	// 2. Claim token during device boot
	app, err := mgr.ClaimZTPToken(token.Token, "SN-RPI5-94812", "London Oxford St Store", ModelRaspberryPi5)
	if err != nil {
		t.Fatalf("ClaimZTPToken failed: %v", err)
	}
	if app.ID != "probe-london-01" {
		t.Errorf("Expected ID probe-london-01, got %s", app.ID)
	}
	if app.Status != "ONLINE" || app.ActiveUplink != UplinkPrimaryWAN {
		t.Errorf("Appliance initial state invalid: status=%s uplink=%s", app.Status, app.ActiveUplink)
	}

	// 3. Re-claiming should fail with ErrTokenClaimed
	_, err = mgr.ClaimZTPToken(token.Token, "SN-RPI5-94812", "London Store", ModelRaspberryPi5)
	if err != ErrTokenClaimed {
		t.Errorf("Expected ErrTokenClaimed, got %v", err)
	}
}

func TestAppliance_CellularFailoverAndRecovery(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	mgr := NewManager(bus)

	var wg sync.WaitGroup
	var failoverReceived, recoveryReceived int
	var mu sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wg.Add(1)
	_, err := bus.Subscribe(ctx, TopicApplianceFailover, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		failoverReceived++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to failover: %v", err)
	}

	wg.Add(1)
	_, err = bus.Subscribe(ctx, TopicApplianceRecovered, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		recoveryReceived++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to recovery: %v", err)
	}

	token, _ := mgr.GenerateZTPToken("tenant-bank", "probe-fra-02", 5*time.Minute)
	app, _ := mgr.ClaimZTPToken(token.Token, "SN-NUC-2104", "Frankfurt Financial Hub", ModelIntelNUCEdge)

	// 1. Simulate Primary WAN fiber cut: primaryWANUp = false, cellularUp = true, RSRP = -82 dBm
	foEvent, err := mgr.ProcessUplinkTelemetry(app.ID, false, true, -82)
	if err != nil {
		t.Fatalf("ProcessUplinkTelemetry failed: %v", err)
	}
	if foEvent == nil || foEvent.NewUplink != UplinkCellular5G {
		t.Fatalf("Expected failover event to CELLULAR_5G, got %+v", foEvent)
	}

	apps := mgr.GetAppliances("tenant-bank")
	if len(apps) != 1 || apps[0].ActiveUplink != UplinkCellular5G || apps[0].Status != "FAILOVER" {
		t.Errorf("Appliance did not transition to FAILOVER state: %+v", apps[0])
	}

	// 2. Simulate Primary WAN link restoration
	recEvent, err := mgr.ProcessUplinkTelemetry(app.ID, true, true, -75)
	if err != nil {
		t.Fatalf("ProcessUplinkTelemetry recovery failed: %v", err)
	}
	if recEvent == nil || recEvent.NewUplink != UplinkPrimaryWAN {
		t.Fatalf("Expected recovery event to PRIMARY_WAN, got %+v", recEvent)
	}

	wg.Wait()

	mu.Lock()
	if failoverReceived != 1 || recoveryReceived != 1 {
		t.Errorf("Expected 1 failover and 1 recovery event, got fo=%d rec=%d", failoverReceived, recoveryReceived)
	}
	mu.Unlock()
}
