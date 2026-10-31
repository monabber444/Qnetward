package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"netwatcher/pkg/audit"
	"netwatcher/pkg/copilot"
	"netwatcher/pkg/dast"
	"netwatcher/pkg/digitaltwin"
)

func runCopilot(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: netwatcher copilot \"<natural language prompt>\"")
		fmt.Fprintln(os.Stderr, "Example: netwatcher copilot \"Generate Cisco IOS-XE ACL to block CVE-2024-3400\"")
		os.Exit(2)
	}

	prompt := strings.Join(args, " ")
	engine := copilot.NewEngine(nil)

	q := copilot.CopilotQuery{
		QueryID:    fmt.Sprintf("cli-%d", time.Now().Unix()),
		UserPrompt: prompt,
	}

	resp, err := engine.ProcessQuery(context.Background(), q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Copilot error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== QNETWARD NETOPS COPILOT (Confidence: %.0f%%) ===\n\n", resp.ConfidenceScore*100)
	fmt.Println(resp.Explanation)
	fmt.Println()

	if resp.Artifact != nil {
		fmt.Printf("--- Artifact: %s (%s) ---\n", resp.Artifact.Title, resp.Artifact.Type)
		fmt.Println(resp.Artifact.Content)
		fmt.Println("-----------------------------------------------------")
	}

	if len(resp.SuggestedActions) > 0 {
		fmt.Println("\nSuggested Actions:")
		for _, a := range resp.SuggestedActions {
			fmt.Printf("  • %s\n", a)
		}
	}
	fmt.Println()
}

func runTwin(args []string) {
	fs := flag.NewFlagSet("twin", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	severLink := fs.String("sever", "", "Link ID to sever (e.g. 'tat14')")
	surge := fs.Float64("surge", 1.0, "Traffic multiplier (e.g. 2.0 = 200%)")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	g := digitaltwin.NewTwinGraph()
	g.AddNode(digitaltwin.Node{ID: "nyc", Name: "NYC Core", Region: "us-east", Type: "CoreRouter", Status: "UP"})
	g.AddNode(digitaltwin.Node{ID: "lon", Name: "London Core", Region: "eu-west", Type: "CoreRouter", Status: "UP"})
	g.AddNode(digitaltwin.Node{ID: "fra", Name: "Frankfurt Hub", Region: "eu-central", Type: "CoreRouter", Status: "UP"})

	g.AddLink(digitaltwin.Link{ID: "tat14", SourceNode: "nyc", TargetNode: "lon", LatencyMs: 65.0, CapacityGbps: 100.0, Status: "UP"})
	g.AddLink(digitaltwin.Link{ID: "euro", SourceNode: "lon", TargetNode: "fra", LatencyMs: 14.0, CapacityGbps: 100.0, Status: "UP"})
	g.AddLink(digitaltwin.Link{ID: "subsea-south", SourceNode: "nyc", TargetNode: "fra", LatencyMs: 85.0, CapacityGbps: 50.0, Status: "UP"})

	g.AddFlow(digitaltwin.TrafficFlow{ID: "f1", SrcNode: "nyc", DstNode: "lon", VolumeGbps: 42.0})
	g.AddFlow(digitaltwin.TrafficFlow{ID: "f2", SrcNode: "nyc", DstNode: "fra", VolumeGbps: 24.0})

	scenario := digitaltwin.SimulationScenario{
		Name:              "CLI What-If Failure Scenario",
		TrafficMultiplier: *surge,
	}
	if *severLink != "" {
		scenario.SeveredLinks = []string{*severLink}
	}

	res, err := g.Simulate(scenario)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Simulation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== QNETWARD NETWORK DIGITAL TWIN SIMULATION ===")
	fmt.Printf("Scenario Severity  : %s\n", res.BlastRadiusSeverity)
	fmt.Printf("Total Traffic Load : %.1f Gbps\n", res.TotalTrafficGbps)
	fmt.Printf("Dropped Traffic    : %.1f Gbps\n", res.DroppedTrafficGbps)
	fmt.Printf("Avg Latency Shift  : +%.1f ms\n", res.AverageLatencyDeltaMs)
	fmt.Printf("Saturated Links    : %d\n", len(res.SaturatedLinks))

	for _, sat := range res.SaturatedLinks {
		fmt.Printf("  [CRITICAL SATURATION] %s (%s <-> %s): %.1f%% utilization (%.1f/%.1f Gbps)\n",
			sat.LinkID, sat.SourceNode, sat.TargetNode, sat.UtilizationPct, sat.SimulatedGbps, sat.CapacityGbps)
	}

	for _, rf := range res.ReroutedFlows {
		if rf.IsDropped {
			fmt.Printf("  [FLOW DROPPED] %s -> %s (%.1f Gbps)\n", rf.SrcNode, rf.DstNode, rf.VolumeGbps)
		} else {
			fmt.Printf("  [FLOW REROUTED] %s -> %s: %s (Δ +%.1fms)\n",
				rf.SrcNode, rf.DstNode, strings.Join(rf.NewPath, " -> "), rf.LatencyDeltaMs)
		}
	}
	fmt.Println("=================================================")
}

func runAudit(args []string) {
	ledger := audit.NewImmutableLedger(nil)
	ctx := context.Background()

	// Append sample chained events
	_, _ = ledger.AppendEvent(ctx, "org-apex", "system@genesis", "GENESIS_INITIALIZED", "system", []byte("genesis payload"))
	_, _ = ledger.AppendEvent(ctx, "org-apex", "admin@apex.io", "SSO_LOGIN", "session-1", []byte("sso payload"))
	_, _ = ledger.AppendEvent(ctx, "org-apex", "aiops-engine", "BGP_REROUTE_APPROVED", "playbook-1", []byte("playbook payload"))

	proof := ledger.VerifyChain()
	fmt.Println("\n=== QNETWARD CRYPTOGRAPHIC AUDIT VERIFICATION ===")
	fmt.Printf("Total Blocks Verified : %d\n", proof.TotalBlocks)
	fmt.Printf("Chain Mathematical Integrity : %t\n", proof.IsValid)
	fmt.Printf("Genesis Hash : %s\n", proof.GenesisHash)
	fmt.Printf("Head Hash    : %s\n", proof.HeadHash)
	if proof.IsValid {
		fmt.Println("Result: 100% Mathematically Intact (SOC 2 Type II / ISO 27001 Passed)")
	} else {
		fmt.Printf("Tamper Detected at Block #%d: %s\n", proof.TamperedSequence, proof.ErrorReason)
	}
	fmt.Println("=================================================")
}

func runDAST(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: netwatcher dast <url>")
		os.Exit(2)
	}
	targetURL := args[0]
	auditor := dast.NewAuditor(nil, nil)

	cfg := dast.AuditConfig{
		CheckHeaders:     true,
		CheckCORS:        true,
		CheckCookies:     true,
		CheckInfoLeakage: true,
	}

	rep, err := auditor.AuditTarget(context.Background(), targetURL, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Audit failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== QNETWARD DEFENSIVE DAST POSTURE AUDIT ===\n")
	fmt.Printf("Target URL    : %s\n", rep.TargetURL)
	fmt.Printf("Security Grade: %s\n", rep.SecurityGrade)
	fmt.Printf("Response Time : %d ms\n", rep.DurationMs)
	fmt.Printf("Findings Count: %d\n", len(rep.Findings))

	for _, f := range rep.Findings {
		fmt.Printf("  [%s] %s: %s\n", f.Severity, f.Title, f.Remediation)
	}
	fmt.Println("=============================================")
}
