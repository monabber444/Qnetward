package copilot

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
)

func TestCopilotEngine_QueryTypes(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	engine := NewEngine(bus)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var busQueryCount int
	var mu sync.Mutex

	wg.Add(3)
	_, err := bus.Subscribe(ctx, TopicCopilotQuery, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		busQueryCount++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to TopicCopilotQuery: %v", err)
	}

	// 1. ACL Generation Query
	q1 := CopilotQuery{
		QueryID:    "q1",
		UserPrompt: "Generate a Cisco IOS-XE ACL to block CVE-2024-3400 traffic",
	}
	r1, err := engine.ProcessQuery(ctx, q1)
	if err != nil {
		t.Fatalf("ProcessQuery q1 failed: %v", err)
	}
	if r1.Artifact == nil || r1.Artifact.Type != ArtifactCiscoACL {
		t.Errorf("Expected Cisco ACL artifact, got %+v", r1.Artifact)
	}
	if !strings.Contains(r1.Artifact.Content, "ip access-list extended") {
		t.Errorf("Expected ACL syntax in content")
	}

	// 2. Post-Mortem Query
	q2 := CopilotQuery{
		QueryID:    "q2",
		UserPrompt: "Draft an executive post-mortem incident report for the Frankfurt outage",
	}
	r2, err := engine.ProcessQuery(ctx, q2)
	if err != nil {
		t.Fatalf("ProcessQuery q2 failed: %v", err)
	}
	if r2.Artifact == nil || r2.Artifact.Type != ArtifactPostMortem {
		t.Errorf("Expected PostMortem artifact, got %+v", r2.Artifact)
	}

	// 3. Diagnostic Query
	q3 := CopilotQuery{
		QueryID:    "q3",
		UserPrompt: "Investigate why packet loss climbed on Singapore router at 14:30",
	}
	r3, err := engine.ProcessQuery(ctx, q3)
	if err != nil {
		t.Fatalf("ProcessQuery q3 failed: %v", err)
	}
	if len(r3.SuggestedActions) == 0 {
		t.Errorf("Expected suggested actions for diagnostic query")
	}

	wg.Wait()

	mu.Lock()
	if busQueryCount != 3 {
		t.Errorf("Expected 3 queries dispatched to eventbus, got %d", busQueryCount)
	}
	mu.Unlock()

	history := engine.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 items in engine history, got %d", len(history))
	}
}
