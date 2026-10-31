package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"netwatcher/drivers/sdk"
	"netwatcher/pkg/eventbus"
)

type TargetItem struct {
	ID          string           `json:"id"`
	TenantID    int64            `json:"tenant_id"`
	ProfileName string           `json:"profile_name"`
	Config      sdk.TargetConfig `json:"config"`
}

type CollectorConfig struct {
	Targets      []TargetItem
	PollInterval time.Duration
	Bus          eventbus.Bus
	Registry     *sdk.ProfileRegistry
	TrapPort     int // UDP trap/syslog port (0 = disabled)
	FlowPort     int // UDP NetFlow port (0 = disabled)
}

type Collector struct {
	cfg      CollectorConfig
	registry *sdk.ProfileRegistry
	bus      eventbus.Bus

	mu       sync.RWMutex
	targets  map[string]TargetItem
	drivers  map[string]sdk.DeviceDriver
	stopCh   chan struct{}
	trapConn *net.UDPConn
	flowConn *net.UDPConn
}

func NewCollector(cfg CollectorConfig) *Collector {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 30 * time.Second
	}
	reg := cfg.Registry
	if reg == nil {
		reg = sdk.DefaultRegistry
	}

	c := &Collector{
		cfg:      cfg,
		registry: reg,
		bus:      cfg.Bus,
		targets:  make(map[string]TargetItem),
		drivers:  make(map[string]sdk.DeviceDriver),
		stopCh:   make(chan struct{}),
	}

	for _, t := range cfg.Targets {
		c.RegisterTarget(t)
	}

	return c
}

func (c *Collector) RegisterTarget(t TargetItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.targets[t.ID] = t

	// Look up or instantiate driver
	if prof, ok := c.registry.Get(t.ProfileName); ok {
		c.drivers[t.ID] = sdk.NewProfileDriver(prof)
	}
}

// PollOnce executes a single telemetry & topology collection round across all registered targets.
func (c *Collector) PollOnce(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for id, target := range c.targets {
		driver, ok := c.drivers[id]
		if !ok {
			slog.Warn("no driver found for target", "id", id, "profile", target.ProfileName)
			continue
		}

		// 1. Collect Metrics
		metrics, err := driver.CollectMetrics(ctx, target.Config)
		if err != nil {
			slog.Error("failed to collect metrics", "target", id, "err", err)
		} else if c.bus != nil {
			payload, _ := json.Marshal(map[string]any{
				"target_id": id,
				"tenant_id": target.TenantID,
				"vendor":    driver.Vendor(),
				"metrics":   metrics,
			})
			_ = c.bus.Publish(ctx, "telemetry.metrics.hardware", payload)
		}

		// 2. Fetch Topology (LLDP/CDP)
		neighbors, err := driver.FetchTopology(ctx, target.Config)
		if err != nil {
			slog.Warn("failed to fetch topology", "target", id, "err", err)
		} else if c.bus != nil && len(neighbors) > 0 {
			payload, _ := json.Marshal(map[string]any{
				"target_id": id,
				"tenant_id": target.TenantID,
				"neighbors": neighbors,
			})
			_ = c.bus.Publish(ctx, "telemetry.topology.lldp", payload)
		}
	}

	return nil
}

// Start begins periodic polling and UDP listeners.
func (c *Collector) Start(ctx context.Context) error {
	if c.cfg.TrapPort > 0 {
		addr := fmt.Sprintf("0.0.0.0:%d", c.cfg.TrapPort)
		uAddr, err := net.ResolveUDPAddr("udp", addr)
		if err == nil {
			conn, err := net.ListenUDP("udp", uAddr)
			if err == nil {
				c.trapConn = conn
				go c.trapListener(conn)
				slog.Info("listening for SNMP/Syslog traps", "port", c.cfg.TrapPort)
			}
		}
	}

	if c.cfg.FlowPort > 0 {
		addr := fmt.Sprintf("0.0.0.0:%d", c.cfg.FlowPort)
		uAddr, err := net.ResolveUDPAddr("udp", addr)
		if err == nil {
			conn, err := net.ListenUDP("udp", uAddr)
			if err == nil {
				c.flowConn = conn
				go c.flowListener(conn)
				slog.Info("listening for NetFlow/IPFIX packets", "port", c.cfg.FlowPort)
			}
		}
	}

	go c.pollLoop(ctx)
	return nil
}

func (c *Collector) Stop() {
	close(c.stopCh)
	if c.trapConn != nil {
		_ = c.trapConn.Close()
	}
	if c.flowConn != nil {
		_ = c.flowConn.Close()
	}
}

func (c *Collector) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.PollInterval)
	defer ticker.Stop()

	// Initial poll
	_ = c.PollOnce(ctx)

	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.PollOnce(ctx)
		}
	}
}

func (c *Collector) trapListener(conn *net.UDPConn) {
	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}
		if n > 0 && c.bus != nil {
			_ = c.bus.Publish(context.Background(), "telemetry.trap", buf[:n])
		}
	}
}

func (c *Collector) flowListener(conn *net.UDPConn) {
	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}
		if n > 0 {
			packet, err := ParseNetFlowV5(buf[:n])
			if err == nil && c.bus != nil {
				data, _ := json.Marshal(packet)
				_ = c.bus.Publish(context.Background(), "telemetry.flow", data)
			}
		}
	}
}
