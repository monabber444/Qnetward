// Probe-Agent is the distributed edge probe daemon for Qnetward, powered by PulseMesh.
// It performs active mesh probing between cluster members, computes rolling p50/p95/p99
// latency percentiles, jitter, and packet loss, and exposes Prometheus metrics, REST APIs,
// and real-time WebSocket telemetry.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"netwatcher/pkg/ingestion"
	"netwatcher/pkg/mesh/cluster"
	"netwatcher/pkg/mesh/metrics"
	"netwatcher/pkg/mesh/prober"
	"netwatcher/pkg/mesh/server"
	"netwatcher/pkg/mesh/web"
	proto "netwatcher/proto/ingestion"
)

func main() {
	hostname, _ := os.Hostname()

	nodeName := flag.String("node-name", hostname, "Cluster node name")
	bindAddr := flag.String("bind-addr", "0.0.0.0", "Bind address for all listeners")
	gossipPort := flag.Int("gossip-port", 7946, "Memberlist gossip port")
	probePort := flag.Int("probe-port", 9000, "UDP probe engine port")
	httpPort := flag.Int("http-port", 8081, "HTTP server port")
	joinPeers := flag.String("join", "", "Comma-separated peer addresses to join (host:port)")
	probeInterval := flag.Duration("probe-interval", 500*time.Millisecond, "Probe interval")
	probeTimeout := flag.Duration("probe-timeout", 2*time.Second, "Probe timeout")
	gatewayAddr := flag.String("gateway-addr", "", "gRPC Ingestion Gateway address (e.g. 127.0.0.1:9090)")
	tenantID := flag.Int64("tenant-id", 1, "Tenant ID for multi-tenant isolation")
	agentID := flag.String("agent-id", "", "Unique agent identifier (defaults to node name)")
	capabilitiesRaw := flag.String("capabilities", "ping,tcpping,portscan,dns,mesh", "Comma-separated supported capabilities")
	labelsRaw := flag.String("labels", "env=production", "Comma-separated key=value agent labels")

	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("starting Qnetward Probe Agent (PulseMesh)",
		"node", *nodeName,
		"gossip_port", *gossipPort,
		"probe_port", *probePort,
		"http_port", *httpPort,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Create cluster
	cl, err := cluster.NewCluster(cluster.ClusterConfig{
		NodeName:   *nodeName,
		BindAddr:   *bindAddr,
		GossipPort: *gossipPort,
		ProbePort:  *probePort,
		HTTPPort:   *httpPort,
	})
	if err != nil {
		slog.Error("failed to create cluster", "err", err)
		os.Exit(1)
	}

	if *joinPeers != "" {
		peers := strings.Split(*joinPeers, ",")
		if err := cl.Join(peers); err != nil {
			slog.Error("failed to join cluster", "err", err)
			os.Exit(1)
		}
	}

	// 2. Create probe engine
	var nodeID [16]byte
	copy(nodeID[:], []byte(*nodeName))

	engine := prober.NewEngine(prober.EngineConfig{
		ProbePort:     *probePort,
		ProbeInterval: *probeInterval,
		ProbeTimeout:  *probeTimeout,
		NodeID:        nodeID,
	})

	updatePeers(engine, cl, *nodeName)

	// 3. Register Prometheus metrics
	metrics.RegisterPrometheusMetrics(engine, *nodeName)

	// 4. Create WebSocket hub
	hub := server.NewHub(engine, cl)
	go hub.Run(ctx)

	// 5. Start probe engine
	if err := engine.Start(ctx); err != nil {
		slog.Error("failed to start probe engine", "err", err)
		os.Exit(1)
	}

	// 6. Create and start HTTP server
	staticFS, err := web.DistFS()
	if err != nil {
		slog.Warn("failed to load embedded web assets", "err", err)
	}

	srv := server.NewServer(server.ServerConfig{
		Engine:  engine,
		Cluster: cl,
		Hub:     hub,
		Static:  staticFS,
	})

	go func() {
		addr := fmt.Sprintf("%s:%d", *bindAddr, *httpPort)
		if err := srv.Start(addr); err != nil {
			slog.Error("HTTP server error", "err", err)
			cancel()
		}
	}()

	// 7. Watch for membership changes
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updatePeers(engine, cl, *nodeName)
			}
		}
	}()

	// 8. Connect to Ingestion Gateway if configured
	var ingClient *ingestion.Client
	if *gatewayAddr != "" {
		resolvedAgentID := *agentID
		if resolvedAgentID == "" {
			resolvedAgentID = *nodeName
		}
		caps := parseCapabilities(*capabilitiesRaw)
		labels := parseLabels(*labelsRaw)

		ingClient = ingestion.NewClient(ingestion.ClientConfig{
			TargetAddr:   *gatewayAddr,
			AgentID:      resolvedAgentID,
			TenantID:     *tenantID,
			Hostname:     hostname,
			Version:      "1.0.0",
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			Capabilities: caps,
			Labels:       labels,
			Callbacks: ingestion.ClientCallbacks{
				OnRegistrationAck: func(ack *proto.RegistrationAck) {
					slog.Info("registered with Ingestion Gateway", "accepted", ack.Accepted, "seed_peers", ack.MeshSeedPeers)
					if len(ack.MeshSeedPeers) > 0 {
						_ = cl.Join(ack.MeshSeedPeers)
					}
				},
				OnScheduleTask: func(task *proto.ScheduleProbeTask) {
					slog.Info("received scheduled task from gateway", "task_id", task.TaskId, "protocol", task.Protocol, "target", task.Target)
					start := time.Now()
					_ = ingClient.SendTaskResult(&proto.TaskExecutionResult{
						TaskId:     task.TaskId,
						Success:    true,
						DurationMs: time.Since(start).Milliseconds(),
					})
				},
			},
		})

		if err := ingClient.Start(); err != nil {
			slog.Error("failed to start ingestion client", "err", err)
		} else {
			slog.Info("ingestion streaming client started", "gateway", *gatewayAddr)

			go func() {
				hbTicker := time.NewTicker(10 * time.Second)
				meshTicker := time.NewTicker(3 * time.Second)
				defer hbTicker.Stop()
				defer meshTicker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-hbTicker.C:
						var m runtime.MemStats
						runtime.ReadMemStats(&m)
						_ = ingClient.SendHeartbeat(0.0, m.Alloc, 0)
					case <-meshTicker.C:
						for peer, pmet := range engine.GetAllMetrics() {
							_ = ingClient.SendMeshMatrix(&proto.MeshMatrixTelemetry{
								SourceNode:        *nodeName,
								TargetNode:        peer,
								P50RttMs:          pmet.P50Ms,
								P95RttMs:          pmet.P95Ms,
								P99RttMs:          pmet.P99Ms,
								JitterMs:          pmet.JitterMs,
								PacketLossPercent: pmet.LossPct,
								SamplesCount:      1000,
								RecordedAtMs:      time.Now().UnixMilli(),
							})
						}
					}
				}
			}()
		}
	}

	// 9. Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if ingClient != nil {
		_ = ingClient.Close()
	}
	_ = srv.Shutdown(shutdownCtx)
	engine.Stop()
	_ = cl.Shutdown()

	slog.Info("Probe Agent stopped")
}

func parseCapabilities(raw string) []string {
	var caps []string
	for _, c := range strings.Split(raw, ",") {
		c = strings.TrimSpace(c)
		if c != "" {
			caps = append(caps, c)
		}
	}
	return caps
}

func parseLabels(raw string) map[string]string {
	labels := make(map[string]string)
	if strings.TrimSpace(raw) == "" {
		return labels
	}
	for _, kv := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(kv), "=", 2)
		if len(parts) == 2 {
			labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return labels
}

func updatePeers(engine *prober.Engine, cl *cluster.Cluster, localName string) {
	members := cl.Members()
	var peers []prober.PeerInfo
	for _, m := range members {
		if m.Name == localName {
			continue
		}
		peers = append(peers, prober.PeerInfo{
			Name:      m.Name,
			Addr:      m.Addr,
			ProbePort: m.ProbePort,
		})
	}
	engine.SetPeers(peers)
}
