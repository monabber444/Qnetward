package digitaltwin

import "time"

// Node represents a physical or logical networking element.
type Node struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Type     string `json:"type"`   // "CoreRouter", "EdgeRouter", "SpineSwitch", "Datacenter"
	Status   string `json:"status"` // "UP", "MAINTENANCE", "DOWN"
}

// Link represents a network cable, circuit, or transit connection between nodes.
type Link struct {
	ID              string  `json:"id"`
	SourceNode      string  `json:"source_node"`
	TargetNode      string  `json:"target_node"`
	LatencyMs       float64 `json:"latency_ms"`
	CapacityGbps    float64 `json:"capacity_gbps"`
	CurrentLoadGbps float64 `json:"current_load_gbps"`
	Status          string  `json:"status"` // "UP", "DEGRADED", "SEVERED"
}

// TrafficFlow represents active workload data flowing between nodes.
type TrafficFlow struct {
	ID         string  `json:"id"`
	SrcNode    string  `json:"src_node"`
	DstNode    string  `json:"dst_node"`
	VolumeGbps float64 `json:"volume_gbps"`
}

// SimulationScenario defines a "What-If" failure experiment.
type SimulationScenario struct {
	Name              string   `json:"name"`
	SeveredLinks      []string `json:"severed_links"`
	MaintenanceNodes  []string `json:"maintenance_nodes"`
	TrafficMultiplier float64  `json:"traffic_multiplier"` // e.g. 1.0 = normal, 2.0 = 200% surge
}

// ReroutedFlow details how an individual traffic flow is displaced during failure.
type ReroutedFlow struct {
	FlowID         string   `json:"flow_id"`
	SrcNode        string   `json:"src_node"`
	DstNode        string   `json:"dst_node"`
	VolumeGbps     float64  `json:"volume_gbps"`
	OriginalPath   []string `json:"original_path"`
	NewPath        []string `json:"new_path"`
	LatencyDeltaMs float64  `json:"latency_delta_ms"`
	IsDropped      bool     `json:"is_dropped"`
}

// SaturatedLink highlights links where simulated traffic exceeds critical capacity.
type SaturatedLink struct {
	LinkID         string  `json:"link_id"`
	SourceNode     string  `json:"source_node"`
	TargetNode     string  `json:"target_node"`
	CapacityGbps   float64 `json:"capacity_gbps"`
	SimulatedGbps  float64 `json:"simulated_gbps"`
	UtilizationPct float64 `json:"utilization_pct"`
}

// SimulationResult captures the blast radius and impact of the scenario.
type SimulationResult struct {
	ScenarioName          string          `json:"scenario_name"`
	TotalTrafficGbps      float64         `json:"total_traffic_gbps"`
	DroppedTrafficGbps    float64         `json:"dropped_traffic_gbps"`
	AffectedFlowsCount    int             `json:"affected_flows_count"`
	ReroutedFlows         []ReroutedFlow  `json:"rerouted_flows"`
	SaturatedLinks        []SaturatedLink `json:"saturated_links"`
	DisconnectedNodes     []string        `json:"disconnected_nodes"`
	AverageLatencyDeltaMs float64         `json:"average_latency_delta_ms"`
	BlastRadiusSeverity   string          `json:"blast_radius_severity"` // "CRITICAL", "HIGH", "MEDIUM", "LOW"
	ExecutedAt            time.Time       `json:"executed_at"`
}
