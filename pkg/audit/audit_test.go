package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
)

func TestImmutableLedger_AppendAndVerify(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	ledger := NewImmutableLedger(bus)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var busEventCount int
	var mu sync.Mutex

	wg.Add(3)
	_, err := bus.Subscribe(ctx, TopicAuditEvent, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		busEventCount++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to TopicAuditEvent: %v", err)
	}

	// 1. Append 3 events
	e1, err := ledger.AppendEvent(ctx, "org-fintech", "user-admin", "SSO_LOGIN", "session-1", []byte(`{"ip":"198.51.100.10"}`))
	if err != nil {
		t.Fatalf("AppendEvent 1 failed: %v", err)
	}
	if e1.Sequence != 0 || e1.PrevHash != GenesisPrevHash {
		t.Errorf("Genesis block misconfigured: seq=%d prev=%s", e1.Sequence, e1.PrevHash)
	}

	e2, err := ledger.AppendEvent(ctx, "org-fintech", "user-admin", "BGP_REROUTE_APPROVED", "bgp-playbook-42", []byte(`{"action":"prepend_as"}`))
	if err != nil {
		t.Fatalf("AppendEvent 2 failed: %v", err)
	}
	if e2.Sequence != 1 || e2.PrevHash != e1.Hash {
		t.Errorf("Block 1 failed to link to genesis: prev=%s want=%s", e2.PrevHash, e1.Hash)
	}

	e3, err := ledger.AppendEvent(ctx, "org-fintech", "aiops-engine", "PORT_QUARANTINED", "switch-port-Te0/1/1", []byte(`{"reason":"flapping"}`))
	if err != nil {
		t.Fatalf("AppendEvent 3 failed: %v", err)
	}
	if e3.Sequence != 2 || e3.PrevHash != e2.Hash {
		t.Errorf("Block 2 failed to link to block 1: prev=%s want=%s", e3.PrevHash, e2.Hash)
	}

	wg.Wait()

	mu.Lock()
	if busEventCount != 3 {
		t.Errorf("Expected 3 events on eventbus, got %d", busEventCount)
	}
	mu.Unlock()

	// 2. Verify intact chain
	proof := ledger.VerifyChain()
	if !proof.IsValid {
		t.Fatalf("Expected valid chain proof, got error: %s (tampered sequence: %d)", proof.ErrorReason, proof.TamperedSequence)
	}
	if proof.TotalBlocks != 3 {
		t.Errorf("Expected 3 total blocks, got %d", proof.TotalBlocks)
	}
}

func TestImmutableLedger_TamperingDetection(t *testing.T) {
	ledger := NewImmutableLedger(nil)
	ctx := context.Background()

	// Append 4 events
	for i := 0; i < 4; i++ {
		_, _ = ledger.AppendEvent(ctx, "org-hospital", "operator", "CONFIG_BACKUP", "backup-id", []byte("sample payload"))
	}

	// Verify before tampering
	proof := ledger.VerifyChain()
	if !proof.IsValid {
		t.Fatalf("Expected chain to be valid before tampering")
	}

	// Simulate malicious in-place modification of Block 1 actorID
	ledger.events[1].ActorID = "attacker-modified"

	// Verify tampering detection
	tamperProof := ledger.VerifyChain()
	if tamperProof.IsValid {
		t.Fatalf("Expected tamper detection to catch modified ActorID, but chain reported valid!")
	}
	if tamperProof.TamperedSequence != 1 {
		t.Errorf("Expected tampered sequence 1, got %d", tamperProof.TamperedSequence)
	}
}
