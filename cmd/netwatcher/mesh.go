package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"netwatcher/pkg/mesh/cluster"
	"netwatcher/pkg/mesh/metrics"
	"netwatcher/pkg/mesh/prober"
	"netwatcher/pkg/mesh/server"
	"netwatcher/pkg/mesh/web"
)

func runMesh(args []string) {
	fs := flag.NewFlagSet("mesh", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	hostname, _ := os.Hostname()

	nodeName := fs.String("node-name", hostname, "Cluster node name")
	bindAddr := fs.String("bind-addr", "0.0.0.0", "Bind address for all listeners")
	gossipPort := fs.Int("gossip-port", 7946, "Memberlist gossip port")
	probePort := fs.Int("probe-port", 9000, "UDP probe engine port")
	httpPort := fs.Int("http-port", 8081, "HTTP server port")
	joinPeers := fs.String("join", "", "Comma-separated peer addresses to join (host:port)")
	probeInterval := fs.Duration("probe-interval", 500*time.Millisecond, "Probe interval")
	probeTimeout := fs.Duration("probe-timeout", 2*time.Second, "Probe timeout")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: netwatcher mesh [flags]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(normalizeArgs(fs, args)); err != nil {
		os.Exit(2)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("starting PulseMesh distributed matrix daemon",
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

	updateMeshPeers(engine, cl, *nodeName)

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
				updateMeshPeers(engine, cl, *nodeName)
			}
		}
	}()

	// 8. Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = srv.Shutdown(shutdownCtx)
	engine.Stop()
	_ = cl.Shutdown()

	slog.Info("PulseMesh stopped")
}

func updateMeshPeers(engine *prober.Engine, cl *cluster.Cluster, localName string) {
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
