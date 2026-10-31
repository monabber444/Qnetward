package ebpf

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestK8sEnricher(t *testing.T) {
	enricher := NewK8sEnricher(1 * time.Minute)

	enricher.RegisterPod("10.244.0.15", K8sMeta{
		PodName:    "payment-service-7f8d",
		Namespace:  "production",
		NodeName:   "worker-node-1",
		Deployment: "payment-service",
		Service:    "payment-svc",
		Labels: map[string]string{
			"app": "payment",
		},
	})

	meta := enricher.Lookup("10.244.0.15")
	if meta == nil {
		t.Fatalf("expected K8s metadata for IP, got nil")
	}
	if meta.PodName != "payment-service-7f8d" || meta.Namespace != "production" {
		t.Errorf("unexpected pod metadata: %+v", meta)
	}

	// Unknown IP lookup
	unknown := enricher.Lookup("192.168.1.1")
	if unknown != nil {
		t.Errorf("expected nil for unknown IP, got %+v", unknown)
	}
}

func TestSocketTracer_IngestionAndEnrichment(t *testing.T) {
	enricher := NewK8sEnricher(1 * time.Minute)
	enricher.RegisterPod("10.244.1.20", K8sMeta{
		PodName:   "web-frontend-5b4c",
		Namespace: "default",
	})
	enricher.RegisterPod("10.244.2.30", K8sMeta{
		PodName:   "database-primary-0",
		Namespace: "database",
	})

	tracer := NewSocketTracer(TracerConfig{
		EnrichKubernetes: true,
		TraceDNS:         true,
		RetransmitAlert:  2,
	}, enricher)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tracer.Start(ctx); err != nil {
		t.Fatalf("failed to start tracer: %v", err)
	}
	defer tracer.Stop()

	// Ingest socket metric
	tracer.IngestRawSocketRecord(SocketMetric{
		Timestamp:   time.Now(),
		SrcIP:       "10.244.1.20",
		SrcPort:     48210,
		DstIP:       "10.244.2.30",
		DstPort:     5432,
		Protocol:    "tcp",
		HandshakeMs: 1.25,
		RttMs:       0.85,
		Retransmits: 3,
		State:       "ESTABLISHED",
	})

	select {
	case m := <-tracer.MetricChannel():
		if m.K8sSource == nil || m.K8sSource.PodName != "web-frontend-5b4c" {
			t.Errorf("expected source to be enriched with K8s pod name, got %+v", m.K8sSource)
		}
		if m.K8sDestination == nil || m.K8sDestination.PodName != "database-primary-0" {
			t.Errorf("expected destination to be enriched with K8s pod name, got %+v", m.K8sDestination)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for socket metric")
	}

	total, retrans, resets := tracer.Stats()
	if total != 1 || retrans != 3 || resets != 0 {
		t.Errorf("unexpected stats: total=%d retrans=%d resets=%d", total, retrans, resets)
	}
}

func TestSocketTracer_ConcurrentSafety(t *testing.T) {
	enricher := NewK8sEnricher(5 * time.Minute)
	tracer := NewSocketTracer(TracerConfig{
		EnrichKubernetes: true,
		TraceDNS:         true,
	}, enricher)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = tracer.Start(ctx)
	defer tracer.Stop()

	var wg sync.WaitGroup
	workers := 10
	iterations := 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tracer.IngestRawSocketRecord(SocketMetric{
					Timestamp:   time.Now(),
					SrcIP:       "10.0.0.1",
					DstIP:       "10.0.0.2",
					Retransmits: 1,
				})
				tracer.IngestRawDNSRecord(DNSMetric{
					Timestamp: time.Now(),
					Question:  "api.internal.net",
					LatencyMs: 0.5,
				})
			}
		}(i)
	}

	wg.Wait()
	total, retrans, _ := tracer.Stats()
	expected := uint64(workers * iterations)
	if total != expected || retrans != expected {
		t.Errorf("expected total=%d, retrans=%d; got total=%d, retrans=%d", expected, expected, total, retrans)
	}
}
