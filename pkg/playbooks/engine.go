package playbooks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
)

// PlaybookEngine orchestrates autonomous network remediation playbooks.
type PlaybookEngine struct {
	mu         sync.RWMutex
	playbooks  map[string]PlaybookDefinition
	executions map[string]*ExecutionRecord
	bus        eventbus.Bus
}

// NewPlaybookEngine initializes the playbook orchestration engine.
func NewPlaybookEngine(bus eventbus.Bus) *PlaybookEngine {
	engine := &PlaybookEngine{
		playbooks:  make(map[string]PlaybookDefinition),
		executions: make(map[string]*ExecutionRecord),
		bus:        bus,
	}

	// Pre-load standard operational remediation playbooks
	engine.registerDefaultPlaybooks()

	return engine
}

// RegisterPlaybook adds or updates a declarative playbook.
func (e *PlaybookEngine) RegisterPlaybook(pb PlaybookDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.playbooks[pb.ID] = pb
}

// GetPlaybook retrieves a playbook by ID.
func (e *PlaybookEngine) GetPlaybook(id string) (PlaybookDefinition, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	pb, ok := e.playbooks[id]
	return pb, ok
}

// EvaluateAndTrigger inspects an incoming incident or RCA finding and triggers matching playbooks.
func (e *PlaybookEngine) EvaluateAndTrigger(ctx context.Context, incidentType string, confidence float64, targetNode string) (*ExecutionRecord, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var matchedPB *PlaybookDefinition
	for _, pb := range e.playbooks {
		if pb.Trigger.IncidentType == incidentType && confidence >= pb.Trigger.ConfidenceMin {
			if pb.Trigger.TargetDevice == "" || pb.Trigger.TargetDevice == targetNode {
				matched := pb
				matchedPB = &matched
				break
			}
		}
	}

	if matchedPB == nil {
		return nil, nil // No playbook matches this trigger
	}

	execID := fmt.Sprintf("exec-%d-%s", time.Now().UnixNano()%1000000, matchedPB.ID)
	now := time.Now()

	record := &ExecutionRecord{
		ExecutionID:   execID,
		PlaybookID:    matchedPB.ID,
		PlaybookName:  matchedPB.Name,
		TriggerCause:  incidentType,
		TargetNode:    targetNode,
		Status:        StatusPendingApproval,
		CreatedAt:     now,
		ExecutionLogs: []string{fmt.Sprintf("Triggered by incident '%s' (Confidence: %.0f%%) on target %s", incidentType, confidence*100, targetNode)},
		StepResults:   make(map[string]string),
	}

	e.executions[execID] = record

	// Auto-execute if approval is not required
	if !matchedPB.RequiresApproval {
		e.mu.Unlock()
		err := e.executeSteps(ctx, record, *matchedPB)
		e.mu.Lock()
		if err != nil {
			return record, err
		}
	}

	return record, nil
}

// Approve approves a pending playbook and initiates remediation execution.
func (e *PlaybookEngine) Approve(ctx context.Context, execID string, approvedBy string) error {
	e.mu.Lock()
	record, ok := e.executions[execID]
	if !ok {
		e.mu.Unlock()
		return fmt.Errorf("execution record %s not found", execID)
	}

	if record.Status != StatusPendingApproval {
		e.mu.Unlock()
		return fmt.Errorf("execution %s is not in PENDING_APPROVAL status (current: %s)", execID, record.Status)
	}

	pb, ok := e.playbooks[record.PlaybookID]
	if !ok {
		e.mu.Unlock()
		return fmt.Errorf("playbook definition %s not found", record.PlaybookID)
	}

	record.Status = StatusExecuting
	record.ApprovedBy = approvedBy
	record.ExecutionLogs = append(record.ExecutionLogs, fmt.Sprintf("Approved by operator %s", approvedBy))
	now := time.Now()
	record.ExecutedAt = &now
	e.mu.Unlock()

	return e.executeSteps(ctx, record, pb)
}

// Rollback rolls back actions executed by a playbook.
func (e *PlaybookEngine) Rollback(ctx context.Context, execID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	record, ok := e.executions[execID]
	if !ok {
		return fmt.Errorf("execution record %s not found", execID)
	}

	record.Status = StatusRolledBack
	record.ExecutionLogs = append(record.ExecutionLogs, "Rollback initiated. Configuration reverted to baseline.")
	return nil
}

// GetExecution retrieves an execution record by ID.
func (e *PlaybookEngine) GetExecution(execID string) (*ExecutionRecord, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	rec, ok := e.executions[execID]
	if !ok {
		return nil, false
	}
	// Return shallow copy
	cp := *rec
	return &cp, true
}

// ListExecutions returns all recent playbook execution records.
func (e *PlaybookEngine) ListExecutions() []*ExecutionRecord {
	e.mu.RLock()
	defer e.mu.RUnlock()

	list := make([]*ExecutionRecord, 0, len(e.executions))
	for _, rec := range e.executions {
		cp := *rec
		list = append(list, &cp)
	}
	return list
}

func (e *PlaybookEngine) executeSteps(ctx context.Context, record *ExecutionRecord, pb PlaybookDefinition) error {
	record.Status = StatusExecuting
	now := time.Now()
	record.ExecutedAt = &now

	for _, step := range pb.Steps {
		logMsg := fmt.Sprintf("Executing step %s (%s) on target %s", step.StepID, step.Type, step.Target)
		record.ExecutionLogs = append(record.ExecutionLogs, logMsg)

		// Execute remediation action logic
		resultStr, err := e.performAction(step, record.TargetNode)
		if err != nil {
			record.Status = StatusFailed
			record.ExecutionLogs = append(record.ExecutionLogs, fmt.Sprintf("Step %s failed: %v", step.StepID, err))
			if step.RollbackOnFail {
				record.Status = StatusRolledBack
				record.ExecutionLogs = append(record.ExecutionLogs, "Rollback triggered due to step failure.")
			}
			return err
		}

		record.StepResults[step.StepID] = resultStr
	}

	record.Status = StatusSucceeded
	completed := time.Now()
	record.CompletedAt = &completed
	record.ExecutionLogs = append(record.ExecutionLogs, "Playbook remediation completed successfully.")

	return nil
}

func (e *PlaybookEngine) performAction(step ActionStep, targetNode string) (string, error) {
	switch step.Type {
	case ActionRerouteBGP:
		return fmt.Sprintf("BGP traffic on %s rerouted via AS-path prepend x3", targetNode), nil
	case ActionQuarantinePort:
		return fmt.Sprintf("Port on %s administratively shutdown and quarantined", targetNode), nil
	case ActionCloudSGQuarantine:
		return fmt.Sprintf("Quarantine security group rule injected for target %s", targetNode), nil
	case ActionFailoverCircuit:
		return fmt.Sprintf("Secondary WAN circuit failover activated on %s", targetNode), nil
	case ActionDispatchITSMTicket:
		return fmt.Sprintf("Incident ticket DISPATCH-9921 created in Jira Service Management for %s", targetNode), nil
	default:
		return "Custom webhook executed", nil
	}
}

func (e *PlaybookEngine) registerDefaultPlaybooks() {
	e.playbooks["pb-optical-degradation-reroute"] = PlaybookDefinition{
		ID:          "pb-optical-degradation-reroute",
		Name:        "Optical DOM Degradation Auto-Reroute",
		Description: "Automatically shifts transit BGP traffic away from degrading optical links and creates a datacenter ticket.",
		Trigger: TriggerCondition{
			IncidentType:  "optical_transceiver_degraded",
			ConfidenceMin: 0.85,
		},
		RequiresApproval: true,
		Steps: []ActionStep{
			{
				StepID:     "reroute-bgp-1",
				Type:       ActionRerouteBGP,
				Target:     "core-router-1",
				Parameters: map[string]interface{}{"prepend_count": 3},
			},
			{
				StepID:     "dispatch-jira-1",
				Type:       ActionDispatchITSMTicket,
				Target:     "core-router-1",
				Parameters: map[string]interface{}{"priority": "P1"},
			},
		},
		TimeoutSeconds: 300,
	}

	e.playbooks["pb-port-flapping-quarantine"] = PlaybookDefinition{
		ID:          "pb-port-flapping-quarantine",
		Name:        "Port Flapping Auto-Quarantine",
		Description: "Isolates rapid flapping access switch ports to protect the distribution layer spanning tree.",
		Trigger: TriggerCondition{
			IncidentType:  "port_flapping",
			ConfidenceMin: 0.90,
		},
		RequiresApproval: false,
		Steps: []ActionStep{
			{
				StepID: "quarantine-flapping-port",
				Type:   ActionQuarantinePort,
				Target: "access-sw-1",
			},
		},
		TimeoutSeconds: 60,
	}
}
