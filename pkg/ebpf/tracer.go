package ebpf

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SocketTracer traces TCP and DNS performance at the kernel socket boundary.
type SocketTracer struct {
	cfg        TracerConfig
	enricher   *K8sEnricher
	metricCh   chan SocketMetric
	dnsCh      chan DNSMetric
	mu         sync.RWMutex
	active     bool
	cancelFunc context.CancelFunc

	// Aggregate statistics
	statsMu        sync.RWMutex
	totalTraced    uint64
	retransmits    uint64
	connectionLoss uint64
}

// NewSocketTracer initializes a new socket tracer engine.
func NewSocketTracer(cfg TracerConfig, enricher *K8sEnricher) *SocketTracer {
	if cfg.RetransmitAlert == 0 {
		cfg.RetransmitAlert = 3
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 1 * time.Second
	}

	return &SocketTracer{
		cfg:      cfg,
		enricher: enricher,
		metricCh: make(chan SocketMetric, 1000),
		dnsCh:    make(chan DNSMetric, 500),
	}
}

// Start begins tracing socket telemetry.
func (t *SocketTracer) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.active {
		t.mu.Unlock()
		return fmt.Errorf("tracer is already active")
	}
	t.active = true
	traceCtx, cancel := context.WithCancel(ctx)
	t.cancelFunc = cancel
	t.mu.Unlock()

	// In production, an eBPF loader loads CO-RE bytecode (sockops/tc).
	// Here, the engine runs the trace processor loop handling kernel telemetry stream.
	go t.runLoop(traceCtx)

	return nil
}

// Stop cleanly stops the socket tracer.
func (t *SocketTracer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return
	}
	t.active = false
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
}

// IngestRawSocketRecord processes an incoming socket event from the kernel hook.
func (t *SocketTracer) IngestRawSocketRecord(m SocketMetric) {
	t.mu.RLock()
	active := t.active
	t.mu.RUnlock()

	if !active {
		return
	}

	// Enrich with Kubernetes metadata if available
	if t.enricher != nil && t.cfg.EnrichKubernetes {
		if m.K8sSource == nil {
			m.K8sSource = t.enricher.Lookup(m.SrcIP)
		}
		if m.K8sDestination == nil {
			m.K8sDestination = t.enricher.Lookup(m.DstIP)
		}
	}

	// Update internal counters
	t.statsMu.Lock()
	t.totalTraced++
	if m.Retransmits > 0 {
		t.retransmits += uint64(m.Retransmits)
	}
	if m.IsReset {
		t.connectionLoss++
	}
	t.statsMu.Unlock()

	// Non-blocking delivery to subscriber channel
	select {
	case t.metricCh <- m:
	default:
		// Drop on channel overflow
	}
}

// IngestRawDNSRecord processes an incoming DNS event from port 53 hooks.
func (t *SocketTracer) IngestRawDNSRecord(d DNSMetric) {
	t.mu.RLock()
	active := t.active
	t.mu.RUnlock()

	if !active || !t.cfg.TraceDNS {
		return
	}

	if t.enricher != nil && t.cfg.EnrichKubernetes && d.K8sSource == nil {
		d.K8sSource = t.enricher.Lookup(d.ClientIP)
	}

	select {
	case t.dnsCh <- d:
	default:
	}
}

// MetricChannel returns the stream of socket metrics.
func (t *SocketTracer) MetricChannel() <-chan SocketMetric {
	return t.metricCh
}

// DNSChannel returns the stream of DNS metrics.
func (t *SocketTracer) DNSChannel() <-chan DNSMetric {
	return t.dnsCh
}

// Stats returns high-level performance counters.
func (t *SocketTracer) Stats() (total uint64, retrans uint64, resets uint64) {
	t.statsMu.RLock()
	defer t.statsMu.RUnlock()
	return t.totalTraced, t.retransmits, t.connectionLoss
}

func (t *SocketTracer) runLoop(ctx context.Context) {
	ticker := time.NewTicker(t.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Cleanup stale K8s caches periodically
			if t.enricher != nil {
				t.enricher.CleanExpired()
			}
		}
	}
}
