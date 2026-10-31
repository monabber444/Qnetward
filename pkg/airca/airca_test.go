package airca_test

import (
	"strings"
	"testing"
	"time"

	"netwatcher/pkg/airca"
	"netwatcher/pkg/topology"
)

func TestAdaptiveBaseline(t *testing.T) {
	baseline := airca.NewAdaptiveBaseline(airca.BaselineConfig{
		Alpha:      0.1,
		ThresholdZ: 3.0,
	})

	// 1. Feed 50 normal samples around 10.0ms (9.5 - 10.5)
	for i := 0; i < 50; i++ {
		val := 10.0 + float64(i%3)*0.2
		zScore, isAnomaly := baseline.Feed(val)
		if isAnomaly {
			t.Fatalf("unexpected anomaly during normal baseline feed: val=%.1f, z=%.2f", val, zScore)
		}
	}

	mean, stdDev, count := baseline.Stats()
	if count != 50 {
		t.Fatalf("expected 50 samples, got %d", count)
	}
	if mean < 9.5 || mean > 11.0 {
		t.Fatalf("mean out of expected range: %.2f", mean)
	}
	if stdDev <= 0 {
		t.Fatalf("expected positive standard deviation, got %.4f", stdDev)
	}

	// 2. Feed a slight normal variation (10.5ms) -> should NOT be an anomaly
	zScore, isAnomaly := baseline.Feed(10.5)
	if isAnomaly {
		t.Fatalf("expected 10.5ms to be within normal range, got z=%.2f", zScore)
	}

	// 3. Feed a dramatic latency spike (150.0ms) -> MUST trigger anomaly (Z >= 3.0)
	zScore, isAnomaly = baseline.Feed(150.0)
	if !isAnomaly {
		t.Fatalf("expected 150.0ms to trigger anomaly, got z=%.2f, isAnomaly=%t", zScore, isAnomaly)
	}
	if zScore < 3.0 {
		t.Fatalf("expected Z-score >= 3.0 for 150ms spike, got %.2f", zScore)
	}
}

func TestRCAEngineAlertStorm(t *testing.T) {
	// Build network topology:
	// core-router-1 -> dist-switch-1 -> web-host-1, web-host-2
	g := topology.NewGraph()
	core := &topology.Node{ID: "core-router-1", Name: "Core Router", Type: topology.NodeCoreRouter}
	dist := &topology.Node{ID: "dist-switch-1", Name: "Dist Switch", Type: topology.NodeDistSwitch}
	h1 := &topology.Node{ID: "web-host-1", Name: "Web Host 1", Type: topology.NodeServer}
	h2 := &topology.Node{ID: "web-host-2", Name: "Web Host 2", Type: topology.NodeServer}

	g.AddNode(core)
	g.AddNode(dist)
	g.AddNode(h1)
	g.AddNode(h2)

	g.AddEdge(topology.Edge{SourceID: "core-router-1", TargetID: "dist-switch-1"})
	g.AddEdge(topology.Edge{SourceID: "dist-switch-1", TargetID: "web-host-1"})
	g.AddEdge(topology.Edge{SourceID: "dist-switch-1", TargetID: "web-host-2"})

	now := time.Now().UTC()
	anomalies := []AnomalyEvent{
		{
			EventID:      "ev-core-1",
			NodeID:       "core-router-1",
			Metric:       "optical_rx_power",
			ObservedVal:  -28.5,
			BaselineMean: -12.0,
			ZScore:       8.4,
			Timestamp:    now,
		},
		{
			EventID:      "ev-dist-1",
			NodeID:       "dist-switch-1",
			Metric:       "packet_loss",
			ObservedVal:  45.0,
			BaselineMean: 0.1,
			ZScore:       6.2,
			Timestamp:    now,
		},
		{
			EventID:      "ev-h1",
			NodeID:       "web-host-1",
			Metric:       "rtt_ms",
			ObservedVal:  250.0,
			BaselineMean: 12.0,
			ZScore:       5.8,
			Timestamp:    now,
		},
		{
			EventID:      "ev-h2",
			NodeID:       "web-host-2",
			Metric:       "rtt_ms",
			ObservedVal:  260.0,
			BaselineMean: 12.0,
			ZScore:       6.0,
			Timestamp:    now,
		},
		{
			EventID:      "ev-core-loss",
			NodeID:       "core-router-1",
			Metric:       "packet_loss",
			ObservedVal:  50.0,
			BaselineMean: 0.0,
			ZScore:       9.1,
			Timestamp:    now,
		},
	}

	rca := airca.NewRCAEngine()
	summary := rca.AnalyzeAlertStorm(anomalies, g)
	if summary == nil {
		t.Fatalf("expected RCA summary")
	}

	if summary.PrimarySuspectNode != "core-router-1" {
		t.Fatalf("expected core-router-1 as primary suspect, got: %s", summary.PrimarySuspectNode)
	}
	if summary.ConfidenceScore < 0.85 {
		t.Fatalf("expected confidence >= 0.85, got %.2f", summary.ConfidenceScore)
	}
	if len(summary.BlastRadiusNodes) != 3 {
		t.Fatalf("expected 3 downstream blast radius nodes, got %d", len(summary.BlastRadiusNodes))
	}
	if !strings.Contains(summary.RootCauseExplanation, "core-router-1") {
		t.Fatalf("explanation does not reference root cause node: %s", summary.RootCauseExplanation)
	}
	if !strings.Contains(summary.SuggestedMitigation, "optical SFP") {
		t.Fatalf("mitigation does not mention optical SFP: %s", summary.SuggestedMitigation)
	}
}

func TestNLQueryTranslator(t *testing.T) {
	translator := airca.NewNLQueryTranslator()

	// 1. Loss & Region Query
	q1 := translator.Translate("show packet loss > 5% in us-east last 2 hours")
	if q1.TargetTable != "probe_results" {
		t.Fatalf("expected target table probe_results, got %s", q1.TargetTable)
	}
	if q1.TimeWindow != 2*time.Hour {
		t.Fatalf("expected 2 hour time window, got %v", q1.TimeWindow)
	}
	if !strings.Contains(q1.SQLWhere, "packet_loss_percent > 5") {
		t.Fatalf("missing packet loss filter in SQL: %s", q1.SQLWhere)
	}
	if !strings.Contains(q1.SQLWhere, "region = 'us-east'") {
		t.Fatalf("missing region filter in SQL: %s", q1.SQLWhere)
	}

	// 2. Alert Query
	q2 := translator.Translate("critical alerts on core-router")
	if q2.TargetTable != "alerts" {
		t.Fatalf("expected alerts table, got %s", q2.TargetTable)
	}
	if !strings.Contains(q2.SQLWhere, "severity = 'critical'") {
		t.Fatalf("missing severity filter in SQL: %s", q2.SQLWhere)
	}
	if !strings.Contains(q2.SQLWhere, "target ILIKE '%core-router%'") {
		t.Fatalf("missing target filter in SQL: %s", q2.SQLWhere)
	}
}

type AnomalyEvent = airca.AnomalyEvent
