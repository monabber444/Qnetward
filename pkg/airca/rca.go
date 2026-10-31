package airca

import (
	"fmt"
	"sort"
	"time"

	"netwatcher/pkg/topology"
)

type AnomalyEvent struct {
	EventID      string    `json:"event_id"`
	NodeID       string    `json:"node_id"`
	Metric       string    `json:"metric"` // e.g. "rtt_ms", "packet_loss", "cpu_percent", "optical_rx_power"
	ObservedVal  float64   `json:"observed_val"`
	BaselineMean float64   `json:"baseline_mean"`
	ZScore       float64   `json:"z_score"`
	Timestamp    time.Time `json:"timestamp"`
}

type RCASummary struct {
	IncidentTitle         string    `json:"incident_title"`
	PrimarySuspectNode    string    `json:"primary_suspect_node"`
	ConfidenceScore       float64   `json:"confidence_score"` // 0.0 to 1.0
	RootCauseExplanation string    `json:"root_cause_explanation"`
	BlastRadiusNodes      []string  `json:"blast_radius_nodes"`
	SuggestedMitigation   string    `json:"suggested_mitigation"`
	GeneratedAt           time.Time `json:"generated_at"`
}

type RCAEngine struct{}

func NewRCAEngine() *RCAEngine {
	return &RCAEngine{}
}

// AnalyzeAlertStorm clusters simultaneous anomaly events and isolates the primary root cause.
func (e *RCAEngine) AnalyzeAlertStorm(events []AnomalyEvent, graph *topology.Graph) *RCASummary {
	if len(events) == 0 {
		return nil
	}

	// 1. Group anomalies by NodeID and find highest Z-score
	nodeScores := make(map[string]float64)
	nodeMetrics := make(map[string][]AnomalyEvent)
	for _, ev := range events {
		nodeScores[ev.NodeID] += ev.ZScore
		nodeMetrics[ev.NodeID] = append(nodeMetrics[ev.NodeID], ev)
	}

	// 2. Identify candidate root cause using topology graph if available
	primarySuspect := ""
	maxScore := -1.0

	if graph != nil {
		// Prefer nodes that have no upstream ancestors in the anomaly set
		for nodeID, score := range nodeScores {
			ancestors := graph.GetUpstreamAncestors(nodeID)
			hasUpstreamAnomaly := false
			for _, anc := range ancestors {
				if _, ok := nodeScores[anc.ID]; ok {
					hasUpstreamAnomaly = true
					break
				}
			}
			// Bonus weight for upstream root nodes
			if !hasUpstreamAnomaly {
				score *= 2.0
			}
			if score > maxScore {
				maxScore = score
				primarySuspect = nodeID
			}
		}
	} else {
		for nodeID, score := range nodeScores {
			if score > maxScore {
				maxScore = score
				primarySuspect = nodeID
			}
		}
	}

	// 3. Compute blast radius
	blastRadiusMap := make(map[string]bool)
	for _, ev := range events {
		if ev.NodeID != primarySuspect {
			blastRadiusMap[ev.NodeID] = true
		}
	}
	if graph != nil && primarySuspect != "" {
		for _, desc := range graph.GetDownstreamDescendants(primarySuspect) {
			blastRadiusMap[desc.ID] = true
		}
	}

	blastRadius := make([]string, 0, len(blastRadiusMap))
	for n := range blastRadiusMap {
		blastRadius = append(blastRadius, n)
	}
	sort.Strings(blastRadius)

	// 4. Formulate explanation and suggested mitigation
	primaryEvents := nodeMetrics[primarySuspect]
	leadMetric := "metric deviation"
	var maxZ float64
	for _, pe := range primaryEvents {
		if pe.ZScore > maxZ {
			maxZ = pe.ZScore
			leadMetric = pe.Metric
		}
	}

	explanation := fmt.Sprintf(
		"Root cause detected on node '%s' showing anomalous %s (Z-score: %.1f). Downstream dependencies experienced cascading service degradation.",
		primarySuspect, leadMetric, maxZ,
	)

	mitigation := "Inspect physical link status, optical SFP power levels, and router transit BGP sessions."
	if leadMetric == "cpu_percent" {
		mitigation = "Review top process thread utilization, investigate potential DDoS packet burst or memory leak."
	} else if leadMetric == "optical_rx_power" {
		mitigation = "Clean fiber patch cord connectors or replace degraded SFP optical transceiver module."
	}

	confidence := 0.85
	if len(events) >= 5 {
		confidence = 0.95
	}

	return &RCASummary{
		IncidentTitle:         fmt.Sprintf("Cascading Outage Rooted in %s", primarySuspect),
		PrimarySuspectNode:    primarySuspect,
		ConfidenceScore:       confidence,
		RootCauseExplanation: explanation,
		BlastRadiusNodes:      blastRadius,
		SuggestedMitigation:   mitigation,
		GeneratedAt:           time.Now().UTC(),
	}
}
