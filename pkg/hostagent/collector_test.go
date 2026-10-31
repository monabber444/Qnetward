package hostagent_test

import (
	"context"
	"testing"
	"time"

	"netwatcher/pkg/hostagent"
)

func TestHostAgentCollection(t *testing.T) {
	col := hostagent.NewCollector()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	telemetry, err := col.Collect(ctx)
	if err != nil {
		t.Fatalf("failed to collect host telemetry: %v", err)
	}

	if telemetry.Hostname == "" {
		t.Fatalf("expected non-empty hostname")
	}
	if telemetry.CPUCores <= 0 {
		t.Fatalf("expected positive CPU cores, got %d", telemetry.CPUCores)
	}
	if telemetry.MemTotalBytes <= 0 {
		t.Fatalf("expected positive total memory, got %d", telemetry.MemTotalBytes)
	}
	if len(telemetry.Disks) == 0 {
		t.Fatalf("expected at least 1 disk partition")
	}
	if len(telemetry.Interfaces) == 0 {
		t.Fatalf("expected at least 1 active interface")
	}
	if len(telemetry.TopProcesses) == 0 {
		t.Fatalf("expected top processes list")
	}
	if len(telemetry.Containers) == 0 {
		t.Fatalf("expected container health status")
	}
	if len(telemetry.SMARTDevices) == 0 {
		t.Fatalf("expected SMART device metrics")
	}
	if telemetry.SMARTDevices[0].HealthStatus != "PASSED" {
		t.Fatalf("unexpected SMART status: %s", telemetry.SMARTDevices[0].HealthStatus)
	}
}
