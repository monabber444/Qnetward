package ingestion

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
	proto "netwatcher/proto/ingestion"
)

func TestIngestionStreamIntegration(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	srv := NewServer(ServerConfig{
		Bus:                      bus,
		HeartbeatIntervalSeconds: 5,
		MeshSeedPeers:            []string{"seed1:7946", "seed2:7946"},
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

	// Track published events on the bus
	var receivedProbe atomicValue
	var receivedHeartbeat atomicValue
	var receivedMesh atomicValue

	ctx := context.Background()
	_, _ = bus.Subscribe(ctx, eventbus.TopicTelemetryProbe, func(c context.Context, msg eventbus.Message) error {
		receivedProbe.Store(string(msg.Data))
		return nil
	})
	_, _ = bus.Subscribe(ctx, eventbus.TopicAgentHeartbeat, func(c context.Context, msg eventbus.Message) error {
		receivedHeartbeat.Store(string(msg.Data))
		return nil
	})
	_, _ = bus.Subscribe(ctx, eventbus.TopicTelemetryMesh, func(c context.Context, msg eventbus.Message) error {
		receivedMesh.Store(string(msg.Data))
		return nil
	})

	// Create and connect edge agent client
	var ackReceived sync.WaitGroup
	ackReceived.Add(1)

	var taskReceived sync.WaitGroup
	taskReceived.Add(1)
	var gotTask *proto.ScheduleProbeTask

	client := NewClient(ClientConfig{
		TargetAddr:   lis.Addr().String(),
		AgentID:      "agent-test-1",
		TenantID:     101,
		Hostname:     "node-edge-1",
		Version:      "v1.0.0",
		OS:           "linux",
		Arch:         "amd64",
		Capabilities: []string{"icmp", "tcp", "mesh"},
		Callbacks: ClientCallbacks{
			OnRegistrationAck: func(ack *proto.RegistrationAck) {
				if ack.Accepted {
					ackReceived.Done()
				}
			},
			OnScheduleTask: func(task *proto.ScheduleProbeTask) {
				gotTask = task
				taskReceived.Done()
			},
		},
	})

	if err := client.Start(); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Close()

	// 1. Verify Registration ACK
	waitWithTimeout(t, &ackReceived, 3*time.Second, "registration ack")

	// Verify server sees connected agent
	agents := srv.ConnectedAgents()
	if len(agents) != 1 {
		t.Fatalf("expected 1 connected agent, got %d", len(agents))
	}
	if agents[0].AgentID != "agent-test-1" || agents[0].TenantID != 101 {
		t.Fatalf("agent info mismatch: %+v", agents[0])
	}

	// 2. Send Heartbeat
	if err := client.SendHeartbeat(15.5, 1024*1024*50, 2); err != nil {
		t.Fatalf("failed to send heartbeat: %v", err)
	}

	waitForCondition(t, 2*time.Second, func() bool {
		return receivedHeartbeat.Load() != ""
	}, "heartbeat event on bus")

	// 3. Send Probe Telemetry
	probeRes := &proto.ProbeTelemetry{
		ProbeId:      "probe-xyz",
		Protocol:     "tcp",
		Target:       "example.com",
		Port:         443,
		Sequence:     1,
		RttMs:        24.5,
		TimestampMs:  time.Now().UnixMilli(),
		ResolvedIp:   "93.184.216.34",
	}
	if err := client.SendProbeResult(probeRes); err != nil {
		t.Fatalf("failed to send probe result: %v", err)
	}

	waitForCondition(t, 2*time.Second, func() bool {
		return receivedProbe.Load() != ""
	}, "probe telemetry event on bus")

	// 4. Send Mesh Matrix Telemetry
	meshMatrix := &proto.MeshMatrixTelemetry{
		SourceNode:         "agent-test-1",
		TargetNode:         "agent-test-2",
		P50RttMs:           1.2,
		P95RttMs:           3.4,
		P99RttMs:           5.6,
		JitterMs:           0.3,
		PacketLossPercent:  0.0,
		SamplesCount:       100,
		RecordedAtMs:       time.Now().UnixMilli(),
	}
	if err := client.SendMeshMatrix(meshMatrix); err != nil {
		t.Fatalf("failed to send mesh matrix: %v", err)
	}

	waitForCondition(t, 2*time.Second, func() bool {
		return receivedMesh.Load() != ""
	}, "mesh matrix event on bus")

	// 5. Dispatch command from server to agent
	cmd := &proto.GatewayCommand{
		TimestampNs: time.Now().UnixNano(),
		CommandId:   "cmd-999",
		Payload: &proto.GatewayCommand_ScheduleTask{
			ScheduleTask: &proto.ScheduleProbeTask{
				TaskId:      "task-ping-1",
				Protocol:    "icmp",
				Target:      "1.1.1.1",
				Count:       4,
				IntervalMs:  500,
				TimeoutMs:   1000,
			},
		},
	}
	if err := srv.SendCommand("agent-test-1", cmd); err != nil {
		t.Fatalf("failed to send command to agent: %v", err)
	}

	waitWithTimeout(t, &taskReceived, 3*time.Second, "scheduled task on agent")
	if gotTask.TaskId != "task-ping-1" || gotTask.Target != "1.1.1.1" {
		t.Fatalf("unexpected task received on client: %+v", gotTask)
	}
}

func waitWithTimeout(t *testing.T, wg *sync.WaitGroup, d time.Duration, name string) {
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

func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool, desc string) {
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

type atomicValue struct {
	mu  sync.RWMutex
	val string
}

func (a *atomicValue) Store(v string) {
	a.mu.Lock()
	a.val = v
	a.mu.Unlock()
}

func (a *atomicValue) Load() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.val
}
