package connector

import (
	"context"
	"net"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
	"netwatcher/pkg/ingestion"
)

func TestConnectorManagerAgentLifecycle(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	srv := ingestion.NewServer(ingestion.ServerConfig{Bus: bus})
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	defer lis.Close()

	go func() {
		_ = srv.Start(lis)
	}()
	defer srv.Stop()

	mgr := NewManager(ManagerConfig{
		Gateway:          srv,
		Bus:              bus,
		HeartbeatTimeout: 200 * time.Millisecond,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = mgr.Start(ctx)
	defer mgr.Stop()

	// 1. Register agents
	mgr.RegisterAgent(ingestion.AgentInfo{
		AgentID:      "agent-a",
		TenantID:     1,
		Hostname:     "node-a",
		Capabilities: []string{"icmp", "tcp"},
	})
	mgr.RegisterAgent(ingestion.AgentInfo{
		AgentID:      "agent-b",
		TenantID:     2,
		Hostname:     "node-b",
		Capabilities: []string{"dns"},
	})

	list := mgr.ListAgents(1)
	if len(list) != 1 || list[0].Info.AgentID != "agent-a" {
		t.Fatalf("expected 1 agent for tenant 1, got %v", list)
	}

	// 2. Select eligible agent
	agent, err := mgr.SelectAgent(1, "icmp")
	if err != nil || agent.Info.AgentID != "agent-a" {
		t.Fatalf("expected agent-a selected, got %v, err=%v", agent, err)
	}

	_, err = mgr.SelectAgent(1, "dns")
	if err != ErrNoEligibleAgents {
		t.Fatalf("expected ErrNoEligibleAgents for DNS on tenant 1, got %v", err)
	}

	// 3. Heartbeat sweep marks stale agent offline
	time.Sleep(300 * time.Millisecond)
	mgr.sweep()

	list = mgr.ListAgents(1)
	if list[0].Status != StatusOffline {
		t.Fatalf("expected status offline after timeout, got %s", list[0].Status)
	}

	// 4. Heartbeat revives agent
	mgr.RecordHeartbeat("agent-a")
	list = mgr.ListAgents(1)
	if list[0].Status != StatusOnline {
		t.Fatalf("expected status online after heartbeat, got %s", list[0].Status)
	}
}
