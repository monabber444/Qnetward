package dast

import (
	"time"
)

// FindingSeverity represents the risk level of an issue.
type FindingSeverity string

const (
	SeverityCritical FindingSeverity = "CRITICAL"
	SeverityHigh     FindingSeverity = "HIGH"
	SeverityMedium   FindingSeverity = "MEDIUM"
	SeverityLow      FindingSeverity = "LOW"
	SeverityInfo     FindingSeverity = "INFO"
)

// DASTFinding describes a discovered security flaw or posture weakness.
type DASTFinding struct {
	ID          string          `json:"id"`
	TargetURL   string          `json:"target_url"`
	Category    string          `json:"category"` // "HEADER_MISSING", "CORS_PERMISSIVE", "COOKIE_INSECURE", "INFO_LEAKAGE", "OAST_CALLBACK"
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Severity    FindingSeverity `json:"severity"`
	Remediation string          `json:"remediation"`
	Evidence    string          `json:"evidence"`
	DetectedAt  time.Time       `json:"detected_at"`
}

// AuditConfig specifies rules and targets for the DAST auditor.
type AuditConfig struct {
	Targets            []string      `json:"targets"`
	Timeout            time.Duration `json:"timeout"`
	CheckHeaders       bool          `json:"check_headers"`
	CheckCORS          bool          `json:"check_cors"`
	CheckCookies       bool          `json:"check_cookies"`
	CheckInfoLeakage   bool          `json:"check_info_leakage"`
	EnableOASTTracking bool          `json:"enable_oast_tracking"`
}

// OASTInteraction represents an out-of-band callback (DNS, HTTP) triggered by an audit.
type OASTInteraction struct {
	Token       string    `json:"token"`
	TargetURL   string    `json:"target_url"`
	Protocol    string    `json:"protocol"` // "DNS", "HTTP"
	RemoteAddr  string    `json:"remote_addr"`
	QueryParam  string    `json:"query_param"`
	Timestamp   time.Time `json:"timestamp"`
	PayloadType string    `json:"payload_type"` // "BLIND_SSRF", "JNDI_LOOKUP", "OUT_OF_BAND_PING"
}

// AuditReport aggregates results for an audited endpoint.
type AuditReport struct {
	TargetURL    string         `json:"target_url"`
	DurationMs   int64          `json:"duration_ms"`
	StatusCode   int            `json:"status_code"`
	Findings     []*DASTFinding `json:"findings"`
	AuditedAt    time.Time      `json:"audited_at"`
	SecurityGrade string         `json:"security_grade"` // "A+", "A", "B", "C", "D", "F"
}
