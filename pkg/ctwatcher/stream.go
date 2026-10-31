package ctwatcher

import (
	"context"
	"encoding/json"
	"sync"

	"netwatcher/pkg/eventbus"
)

const (
	// TopicTyposquatFinding is the eventbus topic for newly detected typosquatting/phishing domains.
	TopicTyposquatFinding = "security.typosquat"
)

// StreamConsumer processes certificate transparency events and triggers alerts.
type StreamConsumer struct {
	detector *BrandDetector
	bus      eventbus.Bus
	mu       sync.RWMutex
	findings []*TyposquatFinding
}

// NewStreamConsumer creates a new Certificate Transparency consumer.
func NewStreamConsumer(detector *BrandDetector, bus eventbus.Bus) *StreamConsumer {
	return &StreamConsumer{
		detector: detector,
		bus:      bus,
		findings: make([]*TyposquatFinding, 0),
	}
}

// ProcessCertificate evaluates a certificate and its SANs, reporting lookalike domains.
func (c *StreamConsumer) ProcessCertificate(ctx context.Context, cert CTCertificate) ([]*TyposquatFinding, error) {
	domains := append([]string{cert.Domain}, cert.SANs...)
	seen := make(map[string]bool)
	var newFindings []*TyposquatFinding

	for _, domain := range domains {
		if domain == "" || seen[domain] {
			continue
		}
		seen[domain] = true

		if finding := c.detector.EvaluateDomain(domain, cert); finding != nil {
			newFindings = append(newFindings, finding)

			c.mu.Lock()
			c.findings = append(c.findings, finding)
			c.mu.Unlock()

			if c.bus != nil {
				data, err := json.Marshal(finding)
				if err == nil {
					_ = c.bus.Publish(ctx, TopicTyposquatFinding, data)
				}
			}
		}
	}

	return newFindings, nil
}

// StartFeed consumes certificates from an input channel until the context is cancelled.
func (c *StreamConsumer) StartFeed(ctx context.Context, feed <-chan CTCertificate) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case cert, ok := <-feed:
			if !ok {
				return nil
			}
			_, _ = c.ProcessCertificate(ctx, cert)
		}
	}
}

// GetFindings returns a snapshot of detected typosquat findings.
func (c *StreamConsumer) GetFindings() []*TyposquatFinding {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make([]*TyposquatFinding, len(c.findings))
	copy(res, c.findings)
	return res
}

// ClearFindings clears the internal findings cache.
func (c *StreamConsumer) ClearFindings() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.findings = make([]*TyposquatFinding, 0)
}
