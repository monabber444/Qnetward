package olap

import (
	"fmt"
	"testing"
	"time"
)

func TestClickHouseEngine_BatchAndPercentiles(t *testing.T) {
	engine := NewEngine()

	// 1. Generate 100 rows with known latency values (1ms to 100ms)
	var batch []*TelemetryRow
	now := time.Now()
	for i := 1; i <= 100; i++ {
		batch = append(batch, &TelemetryRow{
			TenantID:    "tenant-enterprise",
			Timestamp:   now.Add(time.Duration(i) * time.Second),
			DeviceID:    fmt.Sprintf("device-%d", i%5),
			Interface:   "eth0",
			MetricName:  "rtt_ms",
			MetricValue: float64(i),
		})
	}

	stats, err := engine.InsertBatch(batch)
	if err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}
	if stats.RowsCount != 100 {
		t.Errorf("Expected 100 inserted rows, got %d", stats.RowsCount)
	}

	// 2. Query percentiles
	summary, err := engine.QueryPercentiles("tenant-enterprise", "rtt_ms")
	if err != nil {
		t.Fatalf("QueryPercentiles failed: %v", err)
	}
	if summary.Count != 100 {
		t.Errorf("Expected count 100, got %d", summary.Count)
	}
	if summary.Min != 1.0 || summary.Max != 100.0 {
		t.Errorf("Expected min=1.0 max=100.0, got min=%f max=%f", summary.Min, summary.Max)
	}
	if summary.P50 < 49.0 || summary.P50 > 52.0 {
		t.Errorf("Expected P50 ~50ms, got %f", summary.P50)
	}
	if summary.P95 < 94.0 || summary.P95 > 96.0 {
		t.Errorf("Expected P95 ~95ms, got %f", summary.P95)
	}

	// 3. Test S3 Offload
	cutoff := now.Add(51 * time.Second) // Should archive rows 1 through 50
	archiveBytes, archivedCount, err := engine.OffloadToParquetBlock("tenant-enterprise", cutoff)
	if err != nil {
		t.Fatalf("OffloadToParquetBlock failed: %v", err)
	}
	if archivedCount != 50 {
		t.Errorf("Expected 50 archived rows, got %d", archivedCount)
	}
	if len(archiveBytes) == 0 {
		t.Errorf("Expected non-empty archive bytes")
	}
	if engine.TotalRows() != 50 {
		t.Errorf("Expected 50 remaining rows in hot tier, got %d", engine.TotalRows())
	}
}
