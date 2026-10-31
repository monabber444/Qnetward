package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"netwatcher/pkg/hostagent"
)

func main() {
	hostname, _ := os.Hostname()

	agentID := flag.String("agent-id", hostname, "Unique host agent ID")
	tenantID := flag.Int64("tenant-id", 1, "Tenant ID")
	gatewayAddr := flag.String("gateway-addr", "", "gRPC Ingestion Gateway target address")
	pollInterval := flag.Duration("poll-interval", 10*time.Second, "Host OS telemetry collection interval")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("starting Qnetward Host Agent",
		"agent_id", *agentID,
		"tenant_id", *tenantID,
		"poll_interval", *pollInterval,
		"gateway_addr", *gatewayAddr,
	)

	col := hostagent.NewCollector()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initial poll
	if t, err := col.Collect(ctx); err == nil {
		slog.Info("collected host OS telemetry",
			"cores", t.CPUCores,
			"mem_total_gb", float64(t.MemTotalBytes)/(1024*1024*1024),
			"load1", t.LoadAvg1,
			"disks", len(t.Disks),
			"interfaces", len(t.Interfaces),
		)
	}

	go func() {
		ticker := time.NewTicker(*pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if t, err := col.Collect(ctx); err == nil {
					slog.Info("host snapshot",
						"cpu_pct", fmt.Sprintf("%.1f%%", t.CPUPercent),
						"mem_used_pct", fmt.Sprintf("%.1f%%", float64(t.MemUsedBytes)/float64(t.MemTotalBytes)*100.0),
						"load1", t.LoadAvg1,
					)
				}
			}
		}
	}()

	fmt.Println("Qnetward Host Agent is running. Press Ctrl+C to exit.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)
	cancel()

	slog.Info("host agent stopped")
}
