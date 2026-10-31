package audit

import "time"

// AuditEvent represents a single cryptographically chained ledger entry.
type AuditEvent struct {
	Sequence    int64     `json:"sequence"`
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	TenantID    string    `json:"tenant_id"`
	ActorID     string    `json:"actor_id"`
	Action      string    `json:"action"` // e.g. "BGP_REROUTE_APPROVED", "PORT_QUARANTINED", "SSO_LOGIN"
	ResourceID  string    `json:"resource_id"`
	PayloadHash string    `json:"payload_hash"`
	PrevHash    string    `json:"prev_hash"`
	Hash        string    `json:"hash"`
}

// ChainVerificationResult details the mathematical validity of the audit chain.
type ChainVerificationResult struct {
	TotalBlocks      int64     `json:"total_blocks"`
	IsValid          bool      `json:"is_valid"`
	GenesisHash      string    `json:"genesis_hash"`
	HeadHash         string    `json:"head_hash"`
	TamperedSequence int64     `json:"tampered_sequence"` // -1 if chain is intact
	ErrorReason      string    `json:"error_reason"`
	VerifiedAt       time.Time `json:"verified_at"`
}
