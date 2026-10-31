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

	"netwatcher/drivers/sdk"
	"netwatcher/pkg/collector"
	"netwatcher/pkg/eventbus"
)

func main() {
	profilesDir := flag.String("profiles-dir", "drivers/profiles", "Path to declarative YAML device profiles")
	pollInterval := flag.Duration("poll-interval", 30*time.Second, "Hardware polling interval")
	trapPort := flag.Int("trap-port", 1162, "UDP port for SNMP/Syslog traps (0 to disable)")
	flowPort := flag.Int("flow-port", 2055, "UDP port for NetFlow/IPFIX (0 to disable)")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("starting Qnetward Hardware Collector",
		"profiles_dir", *profilesDir,
		"poll_interval", *pollInterval,
		"trap_port", *trapPort,
		"flow_port", *flowPort,
	)

	// Load profiles
	registry := sdk.NewProfileRegistry()
	if _, err := os.Stat(*profilesDir); err == nil {
		if err := registry.LoadFS(os.DirFS(*profilesDir), "."); err != nil {
			slog.Warn("failed to load profiles from filesystem", "err", err)
		} else {
			slog.Info("loaded device profiles", "count", len(registry.List()))
		}
	}

	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	col := collector.NewCollector(collector.CollectorConfig{
		PollInterval: *pollInterval,
		Bus:          bus,
		Registry:     registry,
		TrapPort:     *trapPort,
		FlowPort:     *flowPort,
	})

	if err := col.Start(ctx); err != nil {
		slog.Error("failed to start collector", "err", err)
		os.Exit(1)
	}

	fmt.Println("Qnetward Hardware Collector is running. Press Ctrl+C to exit.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)
	cancel()
	col.Stop()

	slog.Info("collector stopped")
}
