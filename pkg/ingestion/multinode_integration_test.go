package ingestion_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"netwatcher/pkg/connector"
	"netwatcher/pkg/eventbus"
	"netwatcher/pkg/ingestion"
	proto "netwatcher/proto/ingestion"
)

func TestMultiAgent3NodeIntegration(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	srv := ingestion.NewServer(ingestion.ServerConfig{
		Bus:                      bus,
		HeartbeatIntervalSeconds: 5,
		MeshSeedPeers:            []string{"seed.netwatcher.internal:7946"},
	})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	go func() {
		_ = srv.Start(lis)
	}()
	defer srv.Stop()

	// Initialize Connector Manager
	mgr := connector.NewManager(connector.ManagerConfig{
		Gateway:          srv,
		Bus:              bus,
		HeartbeatTimeout: 2 * time.Second,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("failed to start connector manager: %v", err)
	}
	defer mgr.Stop()

	// Track event bus metrics
	var meshEventsCount atomic.Int32
	var heartbeatEventsCount atomic.Int32
	var taskResultsCount atomic.Int32

	_, _ = bus.Subscribe(ctx, eventbus.TopicTelemetryMesh, func(c context.Context, msg eventbus.Message) error {
		meshEventsCount.Add(1)
		return nil
	})
	_, _ = bus.Subscribe(ctx, eventbus.TopicAgentHeartbeat, func(c context.Context, msg eventbus.Message) error {
		heartbeatEventsCount.Add(1)
		return nil
	})
	_, _ = bus.Subscribe(ctx, "task.result", func(c context.Context, msg eventbus.Message) error {
		taskResultsCount.Add(1)
		return nil
	})

	type testNode struct {
		id       string
		tenantID int64
		hostname string
		caps     []string
		client   *ingestion.Client
		ackWg    sync.WaitGroup
		tasksWg  sync.WaitGroup
	}

	nodes := []*testNode{
		{
			id:       "agent-us-east-1",
			tenantID: 10,
			hostname: "node-us-east-1.local",
			caps:     []string{"ping", "tcpping", "mesh"},
		},
		{
			id:       "agent-eu-west-1",
			tenantID: 10,
			hostname: "node-eu-west-1.local",
			caps:     []string{"ping", "tcpping", "portscan", "mesh"},
		},
		{
			id:       "agent-ap-south-1",
			tenantID: 20,
			hostname: "node-ap-south-1.local",
			caps:     []string{"ping", "dns", "mesh"},
		},
	}

	// Connect all 3 nodes concurrently
	for _, n := range nodes {
		n.ackWg.Add(1)
		nodeRef := n
		client := ingestion.NewClient(ingestion.ClientConfig{
			TargetAddr:   lis.Addr().String(),
			AgentID:      n.id,
			TenantID:     n.tenantID,
			Hostname:     n.hostname,
			Version:      "v1.0.0",
			OS:           "linux",
			Arch:         "amd64",
			Capabilities: n.caps,
			Labels: map[string]string{
				"cluster": "prod",
				"region":  n.id[6:],
			},
			Callbacks: ingestion.ClientCallbacks{
				OnRegistrationAck: func(ack *proto.RegistrationAck) {
					if ack.Accepted {
						nodeRef.ackWg.Done()
					}
				},
				OnScheduleTask: func(task *proto.ScheduleProbeTask) {
					nodeRef.tasksWg.Done()
					_ = nodeRef.client.SendTaskResult(&proto.TaskExecutionResult{
						TaskId:     task.TaskId,
						Success:    true,
						DurationMs: 12,
					})
				},
			},
		})
		n.client = client
		if err := client.Start(); err != nil {
			t.Fatalf("failed to start client %s: %v", n.id, err)
		}
		defer client.Close()
	}

	// 1. Verify all 3 agents receive registration ACK
	for _, n := range nodes {
		testWaitWithTimeout(t, &n.ackWg, 3*time.Second, fmt.Sprintf("%s registration ack", n.id))
	}

	// 2. Verify Ingestion Gateway sees all 3 connected agents
	testWaitForCondition(t, 2*time.Second, func() bool {
		return len(srv.ConnectedAgents()) == 3
	}, "3 agents connected to gateway")

	// 3. Verify Connector Manager sees agents filtered by tenant
	testWaitForCondition(t, 2*time.Second, func() bool {
		tenant10Agents := mgr.ListAgents(10)
		tenant20Agents := mgr.ListAgents(20)
		return len(tenant10Agents) == 2 && len(tenant20Agents) == 1
	}, "agents registered in connector manager by tenant")

	// 4. Send Heartbeats from all 3 nodes
	for _, n := range nodes {
		if err := n.client.SendHeartbeat(1.5, 1024*1024*64, 4); err != nil {
			t.Fatalf("failed to send heartbeat from %s: %v", n.id, err)
		}
	}

	testWaitForCondition(t, 2*time.Second, func() bool {
		return heartbeatEventsCount.Load() >= 3
	}, "all 3 heartbeats processed on event bus")

	// 5. Send Mesh Matrix Telemetry from all nodes (simulating 3x3 mesh)
	for i, src := range nodes {
		for j, dst := range nodes {
			if i == j {
				continue
			}
			err := src.client.SendMeshMatrix(&proto.MeshMatrixTelemetry{
				SourceNode:        src.hostname,
				TargetNode:        dst.hostname,
				P50RttMs:          1.2 + float64(i+j),
				P95RttMs:          2.4 + float64(i+j),
				P99RttMs:          3.8 + float64(i+j),
				JitterMs:          0.3,
				PacketLossPercent: 0.0,
				SamplesCount:      1000,
				RecordedAtMs:      time.Now().UnixMilli(),
			})
			if err != nil {
				t.Fatalf("failed to send mesh matrix from %s to %s: %v", src.id, dst.id, err)
			}
		}
	}

	// 3 nodes * 2 peers each = 6 matrix telemetry events
	testWaitForCondition(t, 3*time.Second, func() bool {
		return meshEventsCount.Load() >= 6
	}, "all 6 mesh matrix telemetry events on bus")

	// 6. Test Task Dispatch via Connector Manager
	// Agent in tenant 10 with capability "portscan" should match node-eu-west-1
	nodes[1].tasksWg.Add(1)

	task := &proto.ScheduleProbeTask{
		TaskId:    "task-scan-99",
		Protocol:  "portscan",
		Target:    "192.168.1.1",
		Port:      443,
		TimeoutMs: 1000,
	}

	if err := mgr.DispatchProbeTask(ctx, 10, task); err != nil {
		t.Fatalf("failed to dispatch task to tenant 10: %v", err)
	}

	// Verify target node received and executed task
	testWaitWithTimeout(t, &nodes[1].tasksWg, 3*time.Second, "task delivery to target agent")

	testWaitForCondition(t, 2*time.Second, func() bool {
		return taskResultsCount.Load() >= 1
	}, "task result returned on event bus")
}

func testWaitWithTimeout(t *testing.T, wg *sync.WaitGroup, d time.Duration, name string) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("timed out waiting for %s", name)
	}
}

func testWaitForCondition(t *testing.T, timeout time.Duration, cond func() bool, desc string) {
	t.Helper()
	start := time.Now()
	for time.Since(start) < timeout {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", desc)
}
