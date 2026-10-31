package predictive

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
)

const (
	TopicPredictiveFailure = "predictive.failure"
)

// Engine performs multivariate predictive modeling on hardware telemetry.
type Engine struct {
	mu               sync.RWMutex
	opticalBaselines map[string]*OpticalReading
	predictions      []*FailurePrediction
	bus              eventbus.Bus
}

// NewEngine creates a predictive failure analysis engine.
func NewEngine(bus eventbus.Bus) *Engine {
	return &Engine{
		opticalBaselines: make(map[string]*OpticalReading),
		predictions:      make([]*FailurePrediction, 0),
		bus:              bus,
	}
}

// RecordOpticalReading processes SFP+/QSFP telemetry, checking for laser degradation.
func (e *Engine) RecordOpticalReading(ctx context.Context, r OpticalReading) *FailurePrediction {
	e.mu.Lock()
	defer e.mu.Unlock()

	baseline, exists := e.opticalBaselines[r.InterfaceID]
	if !exists {
		e.opticalBaselines[r.InterfaceID] = &r
		return nil
	}

	deltaTx := r.TxPowerDbm - baseline.TxPowerDbm
	var biasRatio float64
	if baseline.BiasCurrentMa > 0 {
		biasRatio = r.BiasCurrentMa / baseline.BiasCurrentMa
	}

	// Anomaly condition: Tx power drops by >= 2.0 dBm while bias current spikes >= 25%
	if deltaTx <= -2.0 && biasRatio >= 1.25 {
		now := time.Now().UTC()
		pred := &FailurePrediction{
			ID:                 fmt.Sprintf("pred-opt-%d", now.UnixNano()%1000000),
			TargetID:           r.InterfaceID,
			Component:          ComponentOpticalTransceiver,
			Probability:        0.94,
			FailureHorizon:     "24-48h",
			RootCauseIndicator: fmt.Sprintf("Optical Tx power degraded by %.2f dBm with bias current inflated to %.1f%% of baseline", deltaTx, biasRatio*100),
			RecommendedAction:  "Schedule proactive hot-swap replacement of SFP+ transceiver before hard link failure occurs",
			DetectedAt:         now,
		}

		e.predictions = append(e.predictions, pred)

		if e.bus != nil {
			if data, err := json.Marshal(pred); err == nil {
				_ = e.bus.Publish(ctx, TopicPredictiveFailure, data)
			}
		}

		return pred
	}

	return nil
}

// RecordBufferReading evaluates micro-burst rates and buffer headroom trends.
func (e *Engine) RecordBufferReading(ctx context.Context, r BufferReading) *FailurePrediction {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Anomaly condition: Buffer saturation >= 85% with high micro-burst frequency
	if r.IngressUtilPct >= 85.0 && r.MicroBurstsSec >= 50 {
		now := time.Now().UTC()
		pred := &FailurePrediction{
			ID:                 fmt.Sprintf("pred-buf-%d", now.UnixNano()%1000000),
			TargetID:           r.SwitchID,
			Component:          ComponentSwitchBufferPool,
			Probability:        0.89,
			FailureHorizon:     "12-24h",
			RootCauseIndicator: fmt.Sprintf("Ingress buffer headroom depleted (%.1f%% utilized) with %d micro-bursts/sec", r.IngressUtilPct, r.MicroBurstsSec),
			RecommendedAction:  "Enable WRED/ECN congestion notification and expand dynamic QoS buffer allocation to prevent drop storm",
			DetectedAt:         now,
		}

		e.predictions = append(e.predictions, pred)

		if e.bus != nil {
			if data, err := json.Marshal(pred); err == nil {
				_ = e.bus.Publish(ctx, TopicPredictiveFailure, data)
			}
		}

		return pred
	}

	return nil
}

// GetPredictions returns all generated predictive alerts.
func (e *Engine) GetPredictions() []*FailurePrediction {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make([]*FailurePrediction, len(e.predictions))
	copy(res, e.predictions)
	return res
}
