package collector_test

import (
	"context"
	"encoding/binary"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"netwatcher/drivers/sdk"
	"netwatcher/pkg/collector"
	"netwatcher/pkg/eventbus"
)

func TestCollectorPollingAndFlow(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	registry := sdk.NewProfileRegistry()
	ciscoProfile := &sdk.DeviceProfile{
		Name:     "cisco_switch",
		Vendor:   "Cisco",
		Protocol: "snmp_v2c",
	}
	registry.Register(ciscoProfile)

	var hardwareMetricsReceived atomic.Int32
	var lldpEventsReceived atomic.Int32

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, _ = bus.Subscribe(ctx, "telemetry.metrics.hardware", func(c context.Context, msg eventbus.Message) error {
		hardwareMetricsReceived.Add(1)
		return nil
	})
	_, _ = bus.Subscribe(ctx, "telemetry.topology.lldp", func(c context.Context, msg eventbus.Message) error {
		lldpEventsReceived.Add(1)
		return nil
	})

	col := collector.NewCollector(collector.CollectorConfig{
		Targets: []collector.TargetItem{
			{
				ID:          "switch-access-1",
				TenantID:    1,
				ProfileName: "cisco_switch",
				Config: sdk.TargetConfig{
					Host:     "192.168.1.1",
					Port:     161,
					Protocol: "snmp_v2c",
				},
			},
		},
		PollInterval: 10 * time.Second,
		Bus:          bus,
		Registry:     registry,
	})

	// Execute poll
	if err := col.PollOnce(ctx); err != nil {
		t.Fatalf("poll once failed: %v", err)
	}

	waitForCondition(t, 2*time.Second, func() bool {
		return hardwareMetricsReceived.Load() == 1 && lldpEventsReceived.Load() == 1
	}, "hardware metrics and LLDP events received")
}

func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool, desc string) {
	t.Helper()
	start := time.Now()
	for time.Since(start) < timeout {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", desc)
}

func TestNetFlowV5Parsing(t *testing.T) {
	// Construct a synthetic 72-byte NetFlow v5 packet (24 byte header + 1 48-byte record)
	buf := make([]byte, 72)
	binary.BigEndian.PutUint16(buf[0:2], 5)  // Version 5
	binary.BigEndian.PutUint16(buf[2:4], 1)  // Count 1
	binary.BigEndian.PutUint32(buf[4:8], 1000) // SysUptime
	binary.BigEndian.PutUint32(buf[8:12], 1700000000) // UnixSecs
	binary.BigEndian.PutUint32(buf[16:20], 42) // Seq

	// Record at offset 24
	copy(buf[24:28], net.ParseIP("192.168.1.50").To4()) // SrcIP
	copy(buf[28:32], net.ParseIP("8.8.8.8").To4())        // DstIP
	copy(buf[32:36], net.ParseIP("192.168.1.1").To4())  // NextHop
	binary.BigEndian.PutUint16(buf[36:38], 1)            // InputIf
	binary.BigEndian.PutUint16(buf[38:40], 2)            // OutputIf
	binary.BigEndian.PutUint32(buf[40:44], 10)           // Pkts
	binary.BigEndian.PutUint32(buf[44:48], 14600)        // Bytes
	binary.BigEndian.PutUint32(buf[48:52], 100)          // First
	binary.BigEndian.PutUint32(buf[52:56], 500)          // Last
	binary.BigEndian.PutUint16(buf[56:58], 44332)        // SrcPort
	binary.BigEndian.PutUint16(buf[58:60], 53)           // DstPort
	buf[61] = 0x10                                       // TCPFlags (ACK)
	buf[62] = 17                                         // Protocol (UDP)

	packet, err := collector.ParseNetFlowV5(buf)
	if err != nil {
		t.Fatalf("failed to parse NetFlow v5: %v", err)
	}

	if packet.Version != 5 || packet.Count != 1 {
		t.Fatalf("unexpected header: %+v", packet)
	}

	if len(packet.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(packet.Records))
	}

	r := packet.Records[0]
	if r.SrcIP != "192.168.1.50" || r.DstIP != "8.8.8.8" {
		t.Fatalf("unexpected IPs: %s -> %s", r.SrcIP, r.DstIP)
	}
	if r.SrcPort != 44332 || r.DstPort != 53 {
		t.Fatalf("unexpected ports: %d -> %d", r.SrcPort, r.DstPort)
	}
	if r.Packets != 10 || r.Bytes != 14600 {
		t.Fatalf("unexpected packets/bytes: %d, %d", r.Packets, r.Bytes)
	}
	if r.Protocol != 17 {
		t.Fatalf("expected UDP protocol 17, got %d", r.Protocol)
	}
	if r.Duration != 400*time.Millisecond {
		t.Fatalf("expected 400ms duration, got %v", r.Duration)
	}
}
