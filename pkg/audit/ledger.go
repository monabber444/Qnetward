package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
)

const (
	TopicAuditEvent = "audit.event"
	GenesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"
)

// ImmutableLedger implements a SHA-256 hash-chained cryptographic ledger.
type ImmutableLedger struct {
	mu     sync.RWMutex
	events []*AuditEvent
	bus    eventbus.Bus
}

// NewImmutableLedger creates a new tamper-evident audit ledger.
func NewImmutableLedger(bus eventbus.Bus) *ImmutableLedger {
	return &ImmutableLedger{
		events: make([]*AuditEvent, 0),
		bus:    bus,
	}
}

// AppendEvent computes hashes and appends an immutable audit event to the ledger chain.
func (l *ImmutableLedger) AppendEvent(ctx context.Context, tenantID, actorID, action, resourceID string, payload []byte) (*AuditEvent, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	seq := int64(len(l.events))

	// Hash payload
	payloadSha := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(payloadSha[:])

	// Determine prevHash
	prevHash := GenesisPrevHash
	if seq > 0 {
		prevHash = l.events[seq-1].Hash
	}

	// Compute block hash
	hash := computeBlockHash(seq, now, tenantID, actorID, action, resourceID, payloadHash, prevHash)

	event := &AuditEvent{
		Sequence:    seq,
		ID:          fmt.Sprintf("aud-%d-%s", seq, hash[:8]),
		Timestamp:   now,
		TenantID:    tenantID,
		ActorID:     actorID,
		Action:      action,
		ResourceID:  resourceID,
		PayloadHash: payloadHash,
		PrevHash:    prevHash,
		Hash:        hash,
	}

	l.events = append(l.events, event)

	if l.bus != nil {
		if data, err := json.Marshal(event); err == nil {
			_ = l.bus.Publish(ctx, TopicAuditEvent, data)
		}
	}

	return event, nil
}

// VerifyChain verifies the mathematical integrity of every link in the cryptographic chain.
func (l *ImmutableLedger) VerifyChain() *ChainVerificationResult {
	l.mu.RLock()
	defer l.mu.RUnlock()

	total := int64(len(l.events))
	now := time.Now().UTC()

	if total == 0 {
		return &ChainVerificationResult{
			TotalBlocks:      0,
			IsValid:          true,
			GenesisHash:      "",
			HeadHash:         "",
			TamperedSequence: -1,
			VerifiedAt:       now,
		}
	}

	genesisHash := l.events[0].Hash
	headHash := l.events[total-1].Hash

	for i := int64(0); i < total; i++ {
		curr := l.events[i]

		// 1. Verify sequence monotonicity
		if curr.Sequence != i {
			return &ChainVerificationResult{
				TotalBlocks:      total,
				IsValid:          false,
				GenesisHash:      genesisHash,
				HeadHash:         headHash,
				TamperedSequence: i,
				ErrorReason:      fmt.Sprintf("Sequence jump detected at block %d: expected %d", curr.Sequence, i),
				VerifiedAt:       now,
			}
		}

		// 2. Verify prevHash linking
		expectedPrev := GenesisPrevHash
		if i > 0 {
			expectedPrev = l.events[i-1].Hash
		}
		if curr.PrevHash != expectedPrev {
			return &ChainVerificationResult{
				TotalBlocks:      total,
				IsValid:          false,
				GenesisHash:      genesisHash,
				HeadHash:         headHash,
				TamperedSequence: i,
				ErrorReason:      fmt.Sprintf("Broken hash link at block %d: prev_hash %s does not match preceding hash %s", i, curr.PrevHash, expectedPrev),
				VerifiedAt:       now,
			}
		}

		// 3. Verify block hash calculation
		recalculated := computeBlockHash(curr.Sequence, curr.Timestamp, curr.TenantID, curr.ActorID, curr.Action, curr.ResourceID, curr.PayloadHash, curr.PrevHash)
		if recalculated != curr.Hash {
			return &ChainVerificationResult{
				TotalBlocks:      total,
				IsValid:          false,
				GenesisHash:      genesisHash,
				HeadHash:         headHash,
				TamperedSequence: i,
				ErrorReason:      fmt.Sprintf("Hash mismatch at block %d: payload or fields were tampered with", i),
				VerifiedAt:       now,
			}
		}
	}

	return &ChainVerificationResult{
		TotalBlocks:      total,
		IsValid:          true,
		GenesisHash:      genesisHash,
		HeadHash:         headHash,
		TamperedSequence: -1,
		ErrorReason:      "",
		VerifiedAt:       now,
	}
}

// computeBlockHash produces the deterministic SHA-256 hash of all fields in the block.
func computeBlockHash(seq int64, ts time.Time, tenantID, actorID, action, resourceID, payloadHash, prevHash string) string {
	raw := fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s|%s",
		seq,
		ts.Format(time.RFC3339Nano),
		tenantID,
		actorID,
		action,
		resourceID,
		payloadHash,
		prevHash,
	)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// GetEvents returns a slice copy of the most recent events up to limit.
func (l *ImmutableLedger) GetEvents(limit int) []*AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	total := len(l.events)
	if limit <= 0 || limit > total {
		limit = total
	}

	res := make([]*AuditEvent, limit)
	copy(res, l.events[total-limit:])
	return res
}
