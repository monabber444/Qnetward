package predictive

import (
	"context"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
)

func TestPredictiveEngine_OpticalAndBufferPredictions(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	engine := NewEngine(bus)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var busPredictionCount int
	var mu sync.Mutex

	wg.Add(2)
	_, err := bus.Subscribe(ctx, TopicPredictiveFailure, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		busPredictionCount++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to TopicPredictiveFailure: %v", err)
	}

	// 1. Establish optical baseline for core-router interface Te0/1/1
	baseline := OpticalReading{
		InterfaceID:   "core-router-1:Te0/1/1",
		TxPowerDbm:    -2.1,
		RxPowerDbm:    -3.0,
		BiasCurrentMa: 30.0,
		TemperatureC:  38.5,
		Timestamp:     time.Now().Add(-48 * time.Hour),
	}
	p0 := engine.RecordOpticalReading(ctx, baseline)
	if p0 != nil {
		t.Errorf("Baseline reading should not trigger prediction")
	}

	// 2. Transmit degraded reading: Tx drops from -2.1 to -4.5 (-2.4 dBm delta), bias current spikes to 42mA (1.4x)
	degraded := OpticalReading{
		InterfaceID:   "core-router-1:Te0/1/1",
		TxPowerDbm:    -4.5,
		RxPowerDbm:    -3.2,
		BiasCurrentMa: 42.0,
		TemperatureC:  44.0,
		Timestamp:     time.Now(),
	}
	pOptical := engine.RecordOpticalReading(ctx, degraded)
	if pOptical == nil || pOptical.Component != ComponentOpticalTransceiver {
		t.Fatalf("Expected optical transceiver failure prediction, got %+v", pOptical)
	}
	if pOptical.Probability < 0.90 || pOptical.FailureHorizon != "24-48h" {
		t.Errorf("Unexpected probability or horizon: %+v", pOptical)
	}

	// 3. Transmit saturated switch buffer reading
	bufReading := BufferReading{
		SwitchID:       "spine-switch-fra-01",
		IngressUtilPct: 88.5,
		MicroBurstsSec: 64,
		DroppedPackets: 120,
		Timestamp:      time.Now(),
	}
	pBuffer := engine.RecordBufferReading(ctx, bufReading)
	if pBuffer == nil || pBuffer.Component != ComponentSwitchBufferPool {
		t.Fatalf("Expected switch buffer pool failure prediction, got %+v", pBuffer)
	}
	if pBuffer.FailureHorizon != "12-24h" {
		t.Errorf("Expected 12-24h failure horizon, got %s", pBuffer.FailureHorizon)
	}

	wg.Wait()

	mu.Lock()
	if busPredictionCount != 2 {
		t.Errorf("Expected 2 predictions on eventbus, got %d", busPredictionCount)
	}
	mu.Unlock()

	preds := engine.GetPredictions()
	if len(preds) != 2 {
		t.Errorf("Expected 2 stored predictions, got %d", len(preds))
	}
}
