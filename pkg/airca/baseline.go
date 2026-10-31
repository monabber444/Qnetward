package airca

import (
	"math"
	"sync"
)

// AdaptiveBaseline calculates rolling Exponential Moving Average (EMA) and standard deviation
// to compute dynamic Z-scores for anomaly detection.
type AdaptiveBaseline struct {
	mu            sync.RWMutex
	alpha         float64 // Smoothing factor (0.0 < alpha <= 1.0, e.g. 0.1 for responsive tracking)
	thresholdZ    float64 // Anomaly threshold (e.g. 3.0 for 99.7% confidence boundary)
	minVariance   float64 // Floor to prevent division by zero in flat metrics
	warmupSamples uint64  // Initial samples used to calibrate baseline without alerting
	sampleCount   uint64
	mean          float64
	variance      float64
}

type BaselineConfig struct {
	Alpha         float64 // default 0.1
	ThresholdZ    float64 // default 3.0
	MinVariance   float64 // default 0.25
	WarmupSamples uint64  // default 15
}

func NewAdaptiveBaseline(cfg BaselineConfig) *AdaptiveBaseline {
	if cfg.Alpha <= 0 || cfg.Alpha > 1.0 {
		cfg.Alpha = 0.1
	}
	if cfg.ThresholdZ <= 0 {
		cfg.ThresholdZ = 3.0
	}
	if cfg.MinVariance <= 0 {
		cfg.MinVariance = 0.25
	}
	if cfg.WarmupSamples == 0 {
		cfg.WarmupSamples = 15
	}

	return &AdaptiveBaseline{
		alpha:         cfg.Alpha,
		thresholdZ:    cfg.ThresholdZ,
		minVariance:   cfg.MinVariance,
		warmupSamples: cfg.WarmupSamples,
	}
}

// Feed updates the running mean and variance and returns the calculated Z-score and anomaly decision.
func (b *AdaptiveBaseline) Feed(val float64) (zScore float64, isAnomaly bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.sampleCount++

	if b.sampleCount == 1 {
		b.mean = val
		b.variance = b.minVariance
		return 0.0, false
	}

	// Calculate Z-score based on existing baseline before updating
	stdDev := math.Sqrt(math.Max(b.variance, b.minVariance))
	diff := val - b.mean
	zScore = math.Abs(diff) / stdDev

	// Update running EMA mean and variance (Welford-style exponential moving statistics)
	delta := val - b.mean
	b.mean += b.alpha * delta
	b.variance = (1.0 - b.alpha) * (b.variance + b.alpha*delta*delta)

	if b.sampleCount <= b.warmupSamples {
		return zScore, false
	}

	isAnomaly = zScore >= b.thresholdZ
	return zScore, isAnomaly
}

// Stats returns the current mean and standard deviation.
func (b *AdaptiveBaseline) Stats() (mean float64, stdDev float64, count uint64) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.mean, math.Sqrt(math.Max(b.variance, b.minVariance)), b.sampleCount
}
