package digitaltwin

import (
	"math"
	"sync"
	"time"
)

// TwinGraph manages an in-memory topological model of the network for simulation.
type TwinGraph struct {
	mu    sync.RWMutex
	nodes map[string]*Node
	links map[string]*Link
	flows map[string]*TrafficFlow
}

// NewTwinGraph creates a new network digital twin graph.
func NewTwinGraph() *TwinGraph {
	return &TwinGraph{
		nodes: make(map[string]*Node),
		links: make(map[string]*Link),
		flows: make(map[string]*TrafficFlow),
	}
}

// AddNode registers a network node in the digital twin.
func (g *TwinGraph) AddNode(n Node) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[n.ID] = &n
}

// AddLink registers a network link in the digital twin.
func (g *TwinGraph) AddLink(l Link) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.links[l.ID] = &l
}

// AddFlow registers a steady-state traffic flow between nodes.
func (g *TwinGraph) AddFlow(f TrafficFlow) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.flows[f.ID] = &f
}

// Simulate executes a "What-If" failure scenario and calculates the blast radius.
func (g *TwinGraph) Simulate(scenario SimulationScenario) (*SimulationResult, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	multiplier := scenario.TrafficMultiplier
	if multiplier <= 0 {
		multiplier = 1.0
	}

	// 1. Build active topology filter for baseline
	baselineLinks := make(map[string]*Link)
	for id, l := range g.links {
		if l.Status == "UP" {
			baselineLinks[id] = l
		}
	}

	baselineNodes := make(map[string]*Node)
	for id, n := range g.nodes {
		if n.Status == "UP" {
			baselineNodes[id] = n
		}
	}

	// 2. Build mutated topology for simulated scenario
	severedMap := make(map[string]bool)
	for _, id := range scenario.SeveredLinks {
		severedMap[id] = true
	}

	maintMap := make(map[string]bool)
	for _, id := range scenario.MaintenanceNodes {
		maintMap[id] = true
	}

	simLinks := make(map[string]*Link)
	for id, l := range g.links {
		if l.Status == "UP" && !severedMap[id] {
			// Check if endpoints are down
			if !maintMap[l.SourceNode] && !maintMap[l.TargetNode] {
				simLinks[id] = l
			}
		}
	}

	simNodes := make(map[string]*Node)
	for id, n := range g.nodes {
		if n.Status == "UP" && !maintMap[id] {
			simNodes[id] = n
		}
	}

	// 3. Track simulated link loads
	simLinkLoads := make(map[string]float64)
	for id := range simLinks {
		simLinkLoads[id] = 0.0
	}

	var totalTraffic float64
	var droppedTraffic float64
	var reroutedFlows []ReroutedFlow
	var latencyDeltas []float64

	// 4. Recompute paths for all active traffic flows
	for _, flow := range g.flows {
		volume := flow.VolumeGbps * multiplier
		totalTraffic += volume

		// Find baseline path
		basePath, baseLatency, baseFound := findShortestPath(flow.SrcNode, flow.DstNode, baselineNodes, baselineLinks)
		// Find simulated path
		simPath, simLatency, simFound := findShortestPath(flow.SrcNode, flow.DstNode, simNodes, simLinks)

		if !simFound {
			droppedTraffic += volume
			reroutedFlows = append(reroutedFlows, ReroutedFlow{
				FlowID:         flow.ID,
				SrcNode:        flow.SrcNode,
				DstNode:        flow.DstNode,
				VolumeGbps:     volume,
				OriginalPath:   basePath,
				NewPath:        nil,
				LatencyDeltaMs: 0,
				IsDropped:      true,
			})
		} else {
			delta := simLatency - baseLatency
			latencyDeltas = append(latencyDeltas, delta)

			// Record reroute if path changed or latency increased
			if !slicesEqual(basePath, simPath) || !baseFound {
				reroutedFlows = append(reroutedFlows, ReroutedFlow{
					FlowID:         flow.ID,
					SrcNode:        flow.SrcNode,
					DstNode:        flow.DstNode,
					VolumeGbps:     volume,
					OriginalPath:   basePath,
					NewPath:        simPath,
					LatencyDeltaMs: delta,
					IsDropped:      false,
				})
			}

			// Add traffic load to each traversed link in simPath
			for i := 0; i < len(simPath)-1; i++ {
				u, v := simPath[i], simPath[i+1]
				for linkID, l := range simLinks {
					if (l.SourceNode == u && l.TargetNode == v) || (l.SourceNode == v && l.TargetNode == u) {
						simLinkLoads[linkID] += volume
						break
					}
				}
			}
		}
	}

	// 5. Evaluate Saturated Links (> 90% capacity)
	var saturatedLinks []SaturatedLink
	for linkID, load := range simLinkLoads {
		l := simLinks[linkID]
		if l.CapacityGbps > 0 {
			utilization := (load / l.CapacityGbps) * 100.0
			if utilization >= 90.0 {
				saturatedLinks = append(saturatedLinks, SaturatedLink{
					LinkID:         l.ID,
					SourceNode:     l.SourceNode,
					TargetNode:     l.TargetNode,
					CapacityGbps:   l.CapacityGbps,
					SimulatedGbps:  load,
					UtilizationPct: utilization,
				})
			}
		}
	}

	// 6. Compute Disconnected Nodes
	var disconnected []string
	for nodeID := range g.nodes {
		if maintMap[nodeID] {
			disconnected = append(disconnected, nodeID)
			continue
		}
		// Check if node has at least one active incident link
		hasConn := false
		for _, l := range simLinks {
			if l.SourceNode == nodeID || l.TargetNode == nodeID {
				hasConn = true
				break
			}
		}
		if !hasConn {
			disconnected = append(disconnected, nodeID)
		}
	}

	// 7. Calculate Average Latency Shift
	var avgLatencyDelta float64
	if len(latencyDeltas) > 0 {
		var sum float64
		for _, d := range latencyDeltas {
			sum += d
		}
		avgLatencyDelta = sum / float64(len(latencyDeltas))
	}

	// 8. Determine Blast Radius Severity
	severity := "LOW"
	if droppedTraffic > 0 {
		severity = "CRITICAL"
	} else if len(saturatedLinks) > 0 {
		severity = "HIGH"
	} else if len(reroutedFlows) > 0 || avgLatencyDelta > 15.0 {
		severity = "MEDIUM"
	}

	return &SimulationResult{
		ScenarioName:          scenario.Name,
		TotalTrafficGbps:      totalTraffic,
		DroppedTrafficGbps:    droppedTraffic,
		AffectedFlowsCount:    len(reroutedFlows),
		ReroutedFlows:         reroutedFlows,
		SaturatedLinks:        saturatedLinks,
		DisconnectedNodes:     disconnected,
		AverageLatencyDeltaMs: avgLatencyDelta,
		BlastRadiusSeverity:   severity,
		ExecutedAt:            time.Now(),
	}, nil
}

// findShortestPath implements Dijkstra's algorithm to calculate minimum latency paths.
func findShortestPath(src, dst string, nodes map[string]*Node, links map[string]*Link) ([]string, float64, bool) {
	if src == dst {
		return []string{src}, 0, true
	}
	if nodes[src] == nil || nodes[dst] == nil {
		return nil, 0, false
	}

	dist := make(map[string]float64)
	prev := make(map[string]string)
	visited := make(map[string]bool)

	for id := range nodes {
		dist[id] = math.Inf(1)
	}
	dist[src] = 0

	for len(visited) < len(nodes) {
		// Find unvisited node with smallest distance
		var u string
		minDist := math.Inf(1)
		for id := range nodes {
			if !visited[id] && dist[id] < minDist {
				minDist = dist[id]
				u = id
			}
		}

		if u == "" || math.IsInf(minDist, 1) {
			break
		}
		if u == dst {
			break
		}

		visited[u] = true

		// Relax neighbors
		for _, l := range links {
			var v string
			if l.SourceNode == u {
				v = l.TargetNode
			} else if l.TargetNode == u {
				v = l.SourceNode
			} else {
				continue
			}

			if nodes[v] == nil || visited[v] {
				continue
			}

			alt := dist[u] + l.LatencyMs
			if alt < dist[v] {
				dist[v] = alt
				prev[v] = u
			}
		}
	}

	if math.IsInf(dist[dst], 1) {
		return nil, 0, false
	}

	// Reconstruct path
	var path []string
	curr := dst
	for curr != "" {
		path = append([]string{curr}, path...)
		curr = prev[curr]
	}

	return path, dist[dst], true
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// GetTopology returns a snapshot of current nodes and links.
func (g *TwinGraph) GetTopology() ([]*Node, []*Link) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	nodes := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	links := make([]*Link, 0, len(g.links))
	for _, l := range g.links {
		links = append(links, l)
	}
	return nodes, links
}
