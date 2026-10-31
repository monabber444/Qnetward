package olap

import "time"

// TelemetryRow represents a single columnar row in ClickHouse ReplicatedMergeTree.
type TelemetryRow struct {
	TenantID    string            `json:"tenant_id"`
	Timestamp   time.Time         `json:"timestamp"`
	DeviceID    string            `json:"device_id"`
	Interface   string            `json:"interface"`
	MetricName  string            `json:"metric_name"`
	MetricValue float64           `json:"metric_value"`
	Labels      map[string]string `json:"labels"`
}

// PercentileSummary contains aggregated metrics calculated across columnar partitions.
type PercentileSummary struct {
	MetricName string  `json:"metric_name"`
	Count      int64   `json:"count"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Average    float64 `json:"average"`
	P50        float64 `json:"p50"`
	P95        float64 `json:"p95"`
	P99        float64 `json:"p99"`
}

// BatchInsertStats records throughput efficiency for a written block.
type BatchInsertStats struct {
	RowsCount   int64 `json:"rows_count"`
	DurationMs  int64 `json:"duration_ms"`
	BytesTotal  int64 `json:"bytes_total"`
}
