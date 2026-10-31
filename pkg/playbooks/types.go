package playbooks

import "time"

// PlaybookStatus represents the lifecycle state of a remediation execution.
type PlaybookStatus string

const (
	StatusPendingApproval PlaybookStatus = "PENDING_APPROVAL"
	StatusExecuting       PlaybookStatus = "EXECUTING"
	StatusSucceeded       PlaybookStatus = "SUCCEEDED"
	StatusFailed          PlaybookStatus = "FAILED"
	StatusRolledBack      PlaybookStatus = "ROLLED_BACK"
)

// ActionType defines the network remediation operation to execute.
type ActionType string

const (
	ActionRerouteBGP         ActionType = "reroute_bgp"
	ActionQuarantinePort     ActionType = "quarantine_port"
	ActionCloudSGQuarantine  ActionType = "cloud_sg_quarantine"
	ActionFailoverCircuit    ActionType = "failover_circuit"
	ActionDispatchITSMTicket ActionType = "dispatch_ticket"
	ActionCustomWebhook      ActionType = "custom_webhook"
)

// TriggerCondition defines what triggers an automated playbook.
type TriggerCondition struct {
	IncidentType  string  `json:"incident_type"`  // e.g. "optical_transceiver_degraded", "cascading_outage"
	ConfidenceMin float64 `json:"confidence_min"` // e.g. 0.85
	MinSeverity   string  `json:"min_severity"`   // "CRITICAL", "HIGH"
	TargetDevice  string  `json:"target_device"`  // optional target filter or regex
}

// ActionStep defines a single remediation action within a playbook.
type ActionStep struct {
	StepID      string                 `json:"step_id"`
	Type        ActionType             `json:"type"`
	Target      string                 `json:"target"`
	Parameters  map[string]interface{} `json:"parameters"`
	RollbackOnFail bool                `json:"rollback_on_fail"`
}

// PlaybookDefinition represents a declarative remediation workflow.
type PlaybookDefinition struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Trigger        TriggerCondition  `json:"trigger"`
	RequiresApproval bool            `json:"requires_approval"`
	Steps          []ActionStep      `json:"steps"`
	TimeoutSeconds int               `json:"timeout_seconds"`
}

// ExecutionRecord tracks a running or completed playbook instance.
type ExecutionRecord struct {
	ExecutionID   string                 `json:"execution_id"`
	PlaybookID    string                 `json:"playbook_id"`
	PlaybookName  string                 `json:"playbook_name"`
	TriggerCause  string                 `json:"trigger_cause"`
	TargetNode    string                 `json:"target_node"`
	Status        PlaybookStatus         `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	ExecutedAt    *time.Time             `json:"executed_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	ApprovedBy    string                 `json:"approved_by,omitempty"`
	ExecutionLogs []string               `json:"execution_logs"`
	StepResults   map[string]string      `json:"step_results"`
}
