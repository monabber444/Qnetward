package digitaltwin

import (
	"testing"
)

func buildSampleNetwork() *TwinGraph {
	g := NewTwinGraph()

	// Add 3 core nodes: NYC, LON, FRA
	g.AddNode(Node{ID: "nyc", Name: "Core NYC", Region: "us-east", Type: "CoreRouter", Status: "UP"})
	g.AddNode(Node{ID: "lon", Name: "Core London", Region: "eu-west", Type: "CoreRouter", Status: "UP"})
	g.AddNode(Node{ID: "fra", Name: "Core Frankfurt", Region: "eu-central", Type: "CoreRouter", Status: "UP"})

	// Add 3 links
	// Direct Transatlantic link: NYC <-> LON (65ms, 100Gbps)
	g.AddLink(Link{ID: "link-nyc-lon", SourceNode: "nyc", TargetNode: "lon", LatencyMs: 65.0, CapacityGbps: 100.0, Status: "UP"})
	// European backhaul: LON <-> FRA (14ms, 100Gbps)
	g.AddLink(Link{ID: "link-lon-fra", SourceNode: "lon", TargetNode: "fra", LatencyMs: 14.0, CapacityGbps: 100.0, Status: "UP"})
	// Alternative transatlantic link: NYC <-> FRA (85ms, 50Gbps)
	g.AddLink(Link{ID: "link-nyc-fra", SourceNode: "nyc", TargetNode: "fra", LatencyMs: 85.0, CapacityGbps: 50.0, Status: "UP"})

	// Traffic flow: NYC to LON (30Gbps)
	g.AddFlow(TrafficFlow{ID: "flow-nyc-lon", SrcNode: "nyc", DstNode: "lon", VolumeGbps: 30.0})
	// Traffic flow: NYC to FRA (40Gbps)
	g.AddFlow(TrafficFlow{ID: "flow-nyc-fra", SrcNode: "nyc", DstNode: "fra", VolumeGbps: 40.0})

	return g
}

func TestTwinGraph_ShortestPath(t *testing.T) {
	g := buildSampleNetwork()

	nodes, links := g.GetTopology()
	nodeMap := make(map[string]*Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}
	linkMap := make(map[string]*Link)
	for _, l := range links {
		linkMap[l.ID] = l
	}

	// NYC -> LON shortest path should be direct: [nyc, lon] (65ms)
	path, lat, found := findShortestPath("nyc", "lon", nodeMap, linkMap)
	if !found {
		t.Fatalf("Expected path between nyc and lon")
	}
	if len(path) != 2 || path[0] != "nyc" || path[1] != "lon" {
		t.Errorf("Unexpected path: %v", path)
	}
	if lat != 65.0 {
		t.Errorf("Expected 65.0ms, got %f", lat)
	}

	// NYC -> FRA:
	// Via LON: 65 + 14 = 79ms.
	// Direct: 85ms.
	// Dijkstra should choose [nyc, lon, fra] (79ms) over direct (85ms).
	pathFRA, latFRA, foundFRA := findShortestPath("nyc", "fra", nodeMap, linkMap)
	if !foundFRA {
		t.Fatalf("Expected path between nyc and fra")
	}
	if len(pathFRA) != 3 || pathFRA[1] != "lon" {
		t.Errorf("Expected path via lon, got %v with latency %f", pathFRA, latFRA)
	}
	if latFRA != 79.0 {
		t.Errorf("Expected 79.0ms, got %f", latFRA)
	}
}

func TestTwinGraph_SimulateLinkSeverance(t *testing.T) {
	g := buildSampleNetwork()

	// Scenario: Primary Transatlantic link (NYC <-> LON) is severed
	scenario := SimulationScenario{
		Name:              "Sever Primary Transatlantic Cable",
		SeveredLinks:      []string{"link-nyc-lon"},
		TrafficMultiplier: 1.0,
	}

	res, err := g.Simulate(scenario)
	if err != nil {
		t.Fatalf("Simulate failed: %v", err)
	}

	if res.DroppedTrafficGbps > 0 {
		t.Errorf("Traffic should have rerouted via FRA, but got dropped: %f", res.DroppedTrafficGbps)
	}

	if res.AffectedFlowsCount == 0 {
		t.Errorf("Expected affected flows after cutting link-nyc-lon")
	}

	// Rerouted NYC -> LON traffic now goes via FRA: [nyc, fra, lon]
	// NYC -> FRA (85ms) + FRA -> LON (14ms) = 99ms (vs 65ms original direct)
	if res.AverageLatencyDeltaMs <= 0 {
		t.Errorf("Expected latency increase after rerouting via FRA, got %f", res.AverageLatencyDeltaMs)
	}
}

func TestTwinGraph_SimulateSaturationHotspot(t *testing.T) {
	g := buildSampleNetwork()

	// Scenario: Black Friday 2.5x traffic surge + sever link-nyc-lon
	// NYC-FRA link capacity is only 50Gbps.
	// Total flow = (30 + 40) * 1.5 = 105Gbps forced onto NYC-FRA 50Gbps link!
	scenario := SimulationScenario{
		Name:              "Peak Surge with Cable Cut",
		SeveredLinks:      []string{"link-nyc-lon"},
		TrafficMultiplier: 1.5,
	}

	res, err := g.Simulate(scenario)
	if err != nil {
		t.Fatalf("Simulate failed: %v", err)
	}

	if len(res.SaturatedLinks) == 0 {
		t.Fatalf("Expected link saturation on link-nyc-fra, got 0 saturated links")
	}

	sat := res.SaturatedLinks[0]
	if sat.LinkID != "link-nyc-fra" {
		t.Errorf("Expected link-nyc-fra to be saturated, got %s", sat.LinkID)
	}
	if sat.UtilizationPct < 100.0 {
		t.Errorf("Expected utilization > 100%%, got %f%%", sat.UtilizationPct)
	}
	if res.BlastRadiusSeverity != "HIGH" && res.BlastRadiusSeverity != "CRITICAL" {
		t.Errorf("Expected HIGH or CRITICAL severity, got %s", res.BlastRadiusSeverity)
	}
}

func TestTwinGraph_SimulateNodeMaintenance(t *testing.T) {
	g := buildSampleNetwork()

	// Put Frankfurt node into maintenance
	scenario := SimulationScenario{
		Name:             "Frankfurt Maintenance Window",
		MaintenanceNodes: []string{"fra"},
	}

	res, err := g.Simulate(scenario)
	if err != nil {
		t.Fatalf("Simulate failed: %v", err)
	}

	// Flow to FRA (40Gbps) cannot be delivered
	if res.DroppedTrafficGbps != 40.0 {
		t.Errorf("Expected 40.0 Gbps dropped traffic for unreachable FRA node, got %f", res.DroppedTrafficGbps)
	}
	if res.BlastRadiusSeverity != "CRITICAL" {
		t.Errorf("Expected CRITICAL severity for dropped traffic, got %s", res.BlastRadiusSeverity)
	}
}
