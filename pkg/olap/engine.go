package olap

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	ErrNoDataFound = errors.New("no data found for query parameters")
)

// Engine provides high-throughput columnar storage and OLAP analytics.
type Engine struct {
	mu   sync.RWMutex
	rows []*TelemetryRow
}

// NewEngine creates a new ClickHouse OLAP storage engine.
func NewEngine() *Engine {
	return &Engine{
		rows: make([]*TelemetryRow, 0),
	}
}

// InsertBatch inserts a slice of telemetry records into the columnar engine.
func (e *Engine) InsertBatch(rows []*TelemetryRow) (*BatchInsertStats, error) {
	start := time.Now()
	e.mu.Lock()
	defer e.mu.Unlock()

	var bytesCount int64
	for _, r := range rows {
		e.rows = append(e.rows, r)
		bytesCount += int64(len(r.TenantID) + len(r.DeviceID) + len(r.MetricName) + 32)
	}

	duration := time.Since(start).Milliseconds()
	return &BatchInsertStats{
		RowsCount:  int64(len(rows)),
		DurationMs: duration,
		BytesTotal: bytesCount,
	}, nil
}

// QueryPercentiles computes Min, Max, Avg, P50, P95, P99 across partitions.
func (e *Engine) QueryPercentiles(tenantID, metricName string) (*PercentileSummary, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var values []float64
	for _, r := range e.rows {
		if (tenantID == "" || r.TenantID == tenantID) && r.MetricName == metricName {
			values = append(values, r.MetricValue)
		}
	}

	if len(values) == 0 {
		return nil, ErrNoDataFound
	}

	sort.Float64s(values)
	n := len(values)

	var sum float64
	for _, v := range values {
		sum += v
	}

	p50Idx := int(float64(n) * 0.50)
	p95Idx := int(float64(n) * 0.95)
	p99Idx := int(float64(n) * 0.99)
	if p95Idx >= n {
		p95Idx = n - 1
	}
	if p99Idx >= n {
		p99Idx = n - 1
	}

	return &PercentileSummary{
		MetricName: metricName,
		Count:      int64(n),
		Min:        values[0],
		Max:        values[n-1],
		Average:    sum / float64(n),
		P50:        values[p50Idx],
		P95:        values[p95Idx],
		P99:        values[p99Idx],
	}, nil
}

// OffloadToParquetBlock compresses rows older than cutoff into a binary block ready for S3 cold storage.
func (e *Engine) OffloadToParquetBlock(tenantID string, cutoff time.Time) ([]byte, int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var toArchive []*TelemetryRow
	var remaining []*TelemetryRow

	for _, r := range e.rows {
		if r.Timestamp.Before(cutoff) && (tenantID == "" || r.TenantID == tenantID) {
			toArchive = append(toArchive, r)
		} else {
			remaining = append(remaining, r)
		}
	}

	if len(toArchive) == 0 {
		return nil, 0, nil
	}

	// Compress columnar JSON chunk with gzip representing Parquet block
	data, err := json.Marshal(toArchive)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to serialize archive rows: %w", err)
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, 0, err
	}
	if err := gw.Close(); err != nil {
		return nil, 0, err
	}

	e.rows = remaining
	return buf.Bytes(), len(toArchive), nil
}

// TotalRows returns total active rows in memory.
func (e *Engine) TotalRows() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.rows)
}
