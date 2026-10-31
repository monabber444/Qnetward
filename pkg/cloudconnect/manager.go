package cloudconnect

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"netwatcher/pkg/eventbus"
)

// CloudManager coordinates multi-cloud telemetry connectors.
type CloudManager struct {
	mu         sync.RWMutex
	connectors map[string]CloudConnector
	bus        eventbus.Bus
	flowCh     chan VPCFlowRecord

	// Aggregated metrics
	statsMu       sync.RWMutex
	acceptedBytes uint64
	rejectedBytes uint64
	totalPackets  uint64
}

// NewCloudManager creates a multi-cloud coordinator.
func NewCloudManager(bus eventbus.Bus) *CloudManager {
	return &CloudManager{
		connectors: make(map[string]CloudConnector),
		bus:        bus,
		flowCh:     make(chan VPCFlowRecord, 2000),
	}
}

// RegisterConnector registers a cloud connector instance under a tenant/account key.
func (m *CloudManager) RegisterConnector(key string, conn CloudConnector) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectors[key] = conn
}

// GetConnector retrieves a cloud connector by key.
func (m *CloudManager) GetConnector(key string) (CloudConnector, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.connectors[key]
	return c, ok
}

// IngestFlowRecords processes a batch of normalized VPC flow records.
func (m *CloudManager) IngestFlowRecords(records []VPCFlowRecord) {
	var accBytes, rejBytes, pkts uint64

	for _, rec := range records {
		pkts += rec.Packets
		if rec.Action == ActionAccept {
			accBytes += rec.Bytes
		} else {
			rejBytes += rec.Bytes
		}

		// Publish to event bus if configured
		if m.bus != nil {
			topic := fmt.Sprintf("telemetry.cloud.%s.flow", rec.Provider)
			if data, err := json.Marshal(rec); err == nil {
				_ = m.bus.Publish(context.Background(), topic, data)
			}
		}

		select {
		case m.flowCh <- rec:
		default:
		}
	}

	m.statsMu.Lock()
	m.acceptedBytes += accBytes
	m.rejectedBytes += rejBytes
	m.totalPackets += pkts
	m.statsMu.Unlock()
}

// FlowChannel returns the stream of parsed VPC flow records.
func (m *CloudManager) FlowChannel() <-chan VPCFlowRecord {
	return m.flowCh
}

// Stats returns aggregate accepted and rejected traffic metrics.
func (m *CloudManager) Stats() (accepted uint64, rejected uint64, packets uint64) {
	m.statsMu.RLock()
	defer m.statsMu.RUnlock()
	return m.acceptedBytes, m.rejectedBytes, m.totalPackets
}

// PollAllMetrics queries all registered cloud connectors for network telemetry.
func (m *CloudManager) PollAllMetrics(ctx context.Context) ([]CloudNetworkMetric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allMetrics := make([]CloudNetworkMetric, 0, len(m.connectors)*4)
	for key, conn := range m.connectors {
		metrics, err := conn.PollMetrics(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to poll cloud metrics for %s: %w", key, err)
		}
		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, nil
}
