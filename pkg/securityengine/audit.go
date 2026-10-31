package securityengine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type AuditOutcome string

const (
	OutcomeAllowed AuditOutcome = "allowed"
	OutcomeDenied  AuditOutcome = "denied"
	OutcomeBlocked AuditOutcome = "blocked_ssrf"
)

// AuditLogEntry represents an immutable record of every security scanning authorization decision.
type AuditLogEntry struct {
	ID        string       `json:"id"`
	ActorID   string       `json:"actor_id"`
	TenantID  int64        `json:"tenant_id"`
	Target    string       `json:"target"`
	Tier      string       `json:"tier"`
	Action    string       `json:"action"`
	Scope     string       `json:"scope"`
	Outcome   AuditOutcome `json:"outcome"`
	Reason    string       `json:"reason"`
	Timestamp time.Time    `json:"timestamp"`
}

type AuditLogger interface {
	Log(ctx context.Context, entry AuditLogEntry) error
	GetEntries(tenantID int64) []AuditLogEntry
}

// MemoryAuditLogger provides an in-memory thread-safe audit logger.
type MemoryAuditLogger struct {
	mu      sync.RWMutex
	entries []AuditLogEntry
}

func NewMemoryAuditLogger() *MemoryAuditLogger {
	return &MemoryAuditLogger{
		entries: make([]AuditLogEntry, 0),
	}
}

func (l *MemoryAuditLogger) Log(ctx context.Context, entry AuditLogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	l.entries = append(l.entries, entry)
	return nil
}

func (l *MemoryAuditLogger) GetEntries(tenantID int64) []AuditLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var res []AuditLogEntry
	for _, e := range l.entries {
		if tenantID == 0 || e.TenantID == tenantID {
			res = append(res, e)
		}
	}
	return res
}
