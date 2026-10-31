package topology

import (
	"fmt"
	"time"
)

type AlertEvent struct {
	AlertID          string    `json:"alert_id"`
	NodeID           string    `json:"node_id"`
	TenantID         int64     `json:"tenant_id"`
	Metric           string    `json:"metric"` // "icmp_reachability", "interface_down", "latency_high"
	Severity         string    `json:"severity"` // "warning", "critical"
	Message          string    `json:"message"`
	Timestamp        time.Time `json:"timestamp"`
	Suppressed       bool      `json:"suppressed"`
	RootCauseNodeID  string    `json:"root_cause_node_id,omitempty"`
}

type Incident struct {
	ID               string       `json:"id"`
	TenantID         int64        `json:"tenant_id"`
	RootCauseNodeID  string       `json:"root_cause_node_id"`
	RootCauseMessage string       `json:"root_cause_message"`
	RootAlert        AlertEvent   `json:"root_alert"`
	SuppressedAlerts []AlertEvent `json:"suppressed_alerts"`
	CreatedAt        time.Time    `json:"created_at"`
	Status           string       `json:"status"` // "open", "resolved"
}

// SuppressionEngine handles topology-aware upstream alert suppression.
type SuppressionEngine struct{}

func NewSuppressionEngine() *SuppressionEngine {
	return &SuppressionEngine{}
}

// EvaluateAlerts inspects incoming alerts, identifies root causes, suppresses downstream alerts,
// and bundles them into high-level incidents.
func (e *SuppressionEngine) EvaluateAlerts(graph *Graph, rawAlerts []AlertEvent) (active []AlertEvent, incidents []*Incident) {
	if len(rawAlerts) == 0 {
		return nil, nil
	}

	// 1. Identify which nodes are in CRITICAL state
	criticalNodes := make(map[string]AlertEvent)
	for _, alert := range rawAlerts {
		if alert.Severity == "critical" {
			criticalNodes[alert.NodeID] = alert
		}
	}

	// 2. Filter criticalNodes to only TRUE root causes:
	// A node is a root cause if none of its upstream ancestors are in criticalNodes.
	rootCauses := make(map[string]AlertEvent)
	for nodeID, alert := range criticalNodes {
		hasCriticalAncestor := false
		for _, ancestor := range graph.GetUpstreamAncestors(nodeID) {
			if _, ok := criticalNodes[ancestor.ID]; ok {
				hasCriticalAncestor = true
				break
			}
		}
		if !hasCriticalAncestor {
			rootCauses[nodeID] = alert
		}
	}

	// 3. For each true root cause, suppress all its downstream descendants
	// downstreamNodeID -> upstreamRootAlert
	suppressionMap := make(map[string]AlertEvent)
	for rootID, rootAlert := range rootCauses {
		for _, desc := range graph.GetDownstreamDescendants(rootID) {
			suppressionMap[desc.ID] = rootAlert
		}
	}

	// 4. Partition alerts into active vs suppressed, grouping by root cause incident
	incidentMap := make(map[string]*Incident)

	for _, alert := range rawAlerts {
		if rootAlert, suppressed := suppressionMap[alert.NodeID]; suppressed && alert.NodeID != rootAlert.NodeID {
			// Downstream alert is suppressed!
			alert.Suppressed = true
			alert.RootCauseNodeID = rootAlert.NodeID

			inc, ok := incidentMap[rootAlert.NodeID]
			if !ok {
				inc = &Incident{
					ID:               fmt.Sprintf("inc-%s-%d", rootAlert.NodeID, time.Now().UnixNano()),
					TenantID:         rootAlert.TenantID,
					RootCauseNodeID:  rootAlert.NodeID,
					RootCauseMessage: fmt.Sprintf("Root Cause Failure on %s: %s", rootAlert.NodeID, rootAlert.Message),
					RootAlert:        rootAlert,
					SuppressedAlerts: make([]AlertEvent, 0),
					CreatedAt:        time.Now().UTC(),
					Status:           "open",
				}
				incidentMap[rootAlert.NodeID] = inc
			}
			inc.SuppressedAlerts = append(inc.SuppressedAlerts, alert)
		} else {
			// Root cause or independent alert remains active
			active = append(active, alert)
		}
	}

	for _, inc := range incidentMap {
		incidents = append(incidents, inc)
	}

	return active, incidents
}
