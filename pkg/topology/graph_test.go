package topology_test

import (
	"testing"
	"time"

	"netwatcher/drivers/sdk"
	"netwatcher/pkg/topology"
)

func TestTopologyGraphAndLLDP(t *testing.T) {
	g := topology.NewGraph()

	// 1. Add Core Router
	core := &topology.Node{
		ID:     "core-router-1",
		Name:   "Core Router 01",
		Type:   topology.NodeCoreRouter,
		Status: topology.StatusUp,
		IP:     "10.0.0.1",
	}
	g.AddNode(core)

	// 2. Discover Distribution Switch via LLDP
	lldpNeighbors := []sdk.LLDPNeighbor{
		{
			LocalPort:          "TenGigabitEthernet0/1/1",
			RemoteChassisID:    "00:50:56:a1:b2:c3",
			RemotePort:         "TenGigabitEthernet1/1/1",
			RemoteSystemName:   "dist-switch-01",
			RemoteManagementIP: "10.0.1.1",
		},
	}
	g.BuildFromLLDP(core.ID, lldpNeighbors)

	// Verify dist-switch-01 was added
	distNode, ok := g.GetNode("dist-switch-01")
	if !ok {
		t.Fatalf("dist-switch-01 node not created from LLDP")
	}
	if distNode.IP != "10.0.1.1" {
		t.Fatalf("unexpected IP: %s", distNode.IP)
	}

	// Verify parent/child relationship
	children := g.GetChildren(core.ID)
	if len(children) != 1 || children[0].ID != "dist-switch-01" {
		t.Fatalf("expected dist-switch-01 as child of core, got %+v", children)
	}
	parents := g.GetParents("dist-switch-01")
	if len(parents) != 1 || parents[0].ID != "core-router-1" {
		t.Fatalf("expected core-router-1 as parent of dist, got %+v", parents)
	}
}

func TestDependencyAwareUpstreamSuppression(t *testing.T) {
	// Build a 4-tier network hierarchy:
	// Core Router (CR-1)
	//   -> Dist Switch (DS-1)
	//        -> Access Switch (AS-1)
	//             -> Web Server (WS-1)
	//             -> DB Server (DB-1)
	g := topology.NewGraph()

	cr := &topology.Node{ID: "CR-1", Name: "Core Router", Type: topology.NodeCoreRouter, Status: topology.StatusDown}
	ds := &topology.Node{ID: "DS-1", Name: "Dist Switch", Type: topology.NodeDistSwitch, Status: topology.StatusDown}
	as := &topology.Node{ID: "AS-1", Name: "Access Switch", Type: topology.NodeAccessSwitch, Status: topology.StatusDown}
	ws := &topology.Node{ID: "WS-1", Name: "Web Server", Type: topology.NodeServer, Status: topology.StatusDown}
	db := &topology.Node{ID: "DB-1", Name: "DB Server", Type: topology.NodeServer, Status: topology.StatusDown}

	// Separate independent branch
	cr2 := &topology.Node{ID: "CR-2", Name: "Backup Router", Type: topology.NodeCoreRouter, Status: topology.StatusUp}

	g.AddNode(cr)
	g.AddNode(ds)
	g.AddNode(as)
	g.AddNode(ws)
	g.AddNode(db)
	g.AddNode(cr2)

	g.AddEdge(topology.Edge{SourceID: "CR-1", TargetID: "DS-1", LinkType: "lldp"})
	g.AddEdge(topology.Edge{SourceID: "DS-1", TargetID: "AS-1", LinkType: "lldp"})
	g.AddEdge(topology.Edge{SourceID: "AS-1", TargetID: "WS-1", LinkType: "gateway"})
	g.AddEdge(topology.Edge{SourceID: "AS-1", TargetID: "DB-1", LinkType: "gateway"})

	// Verify transitive descendants of CR-1
	descendants := g.GetDownstreamDescendants("CR-1")
	if len(descendants) != 4 {
		t.Fatalf("expected 4 downstream descendants of CR-1, got %d", len(descendants))
	}

	// Simulate an alert storm:
	// CR-1 goes down. Consequently, DS-1, AS-1, WS-1, and DB-1 all fail reachability checks.
	now := time.Now().UTC()
	rawAlerts := []topology.AlertEvent{
		{
			AlertID:   "alert-cr1",
			NodeID:    "CR-1",
			TenantID:  1,
			Metric:    "icmp_reachability",
			Severity:  "critical",
			Message:   "BGP transit peer down, 100% packet loss",
			Timestamp: now,
		},
		{
			AlertID:   "alert-ds1",
			NodeID:    "DS-1",
			TenantID:  1,
			Metric:    "icmp_reachability",
			Severity:  "critical",
			Message:   "Host unreachable",
			Timestamp: now,
		},
		{
			AlertID:   "alert-as1",
			NodeID:    "AS-1",
			TenantID:  1,
			Metric:    "icmp_reachability",
			Severity:  "critical",
			Message:   "Host unreachable",
			Timestamp: now,
		},
		{
			AlertID:   "alert-ws1",
			NodeID:    "WS-1",
			TenantID:  1,
			Metric:    "http_synthetic",
			Severity:  "critical",
			Message:   "Connection timed out",
			Timestamp: now,
		},
		{
			AlertID:   "alert-db1",
			NodeID:    "DB-1",
			TenantID:  1,
			Metric:    "port_scan",
			Severity:  "critical",
			Message:   "Port 5432 closed or unreachable",
			Timestamp: now,
		},
		// Independent warning on CR-2 that should NOT be suppressed
		{
			AlertID:   "alert-cr2",
			NodeID:    "CR-2",
			TenantID:  1,
			Metric:    "temperature_celsius",
			Severity:  "warning",
			Message:   "Chassis temperature 44C exceeds 40C warning threshold",
			Timestamp: now,
		},
	}

	engine := topology.NewSuppressionEngine()
	active, incidents := engine.EvaluateAlerts(g, rawAlerts)

	// 1. Check active alerts
	// Only CR-1 (root cause critical) and CR-2 (independent warning) should be active
	if len(active) != 2 {
		t.Fatalf("expected 2 active alerts, got %d", len(active))
	}

	activeNodes := map[string]bool{active[0].NodeID: true, active[1].NodeID: true}
	if !activeNodes["CR-1"] || !activeNodes["CR-2"] {
		t.Fatalf("unexpected active alerts: %+v", active)
	}

	// 2. Check Incident generation
	if len(incidents) != 1 {
		t.Fatalf("expected 1 grouped incident, got %d", len(incidents))
	}

	inc := incidents[0]
	if inc.RootCauseNodeID != "CR-1" {
		t.Fatalf("expected root cause node CR-1, got %s", inc.RootCauseNodeID)
	}

	// 4 downstream alerts should be suppressed inside this incident
	if len(inc.SuppressedAlerts) != 4 {
		t.Fatalf("expected 4 suppressed alerts inside incident, got %d", len(inc.SuppressedAlerts))
	}

	for _, supp := range inc.SuppressedAlerts {
		if !supp.Suppressed {
			t.Fatalf("alert %s not marked suppressed", supp.AlertID)
		}
		if supp.RootCauseNodeID != "CR-1" {
			t.Fatalf("alert %s has incorrect root cause %s", supp.AlertID, supp.RootCauseNodeID)
		}
	}
}
