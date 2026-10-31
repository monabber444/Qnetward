package playbooks

import (
	"context"
	"testing"
)

func TestPlaybookEngine_ApprovalAndExecution(t *testing.T) {
	engine := NewPlaybookEngine(nil)

	// Trigger optical DOM playbook (requires approval)
	exec, err := engine.EvaluateAndTrigger(context.Background(), "optical_transceiver_degraded", 0.95, "core-router-1")
	if err != nil {
		t.Fatalf("unexpected trigger error: %v", err)
	}
	if exec == nil {
		t.Fatalf("expected playbook execution, got nil")
	}

	if exec.Status != StatusPendingApproval {
		t.Errorf("expected status PENDING_APPROVAL, got %s", exec.Status)
	}

	// Approve execution
	err = engine.Approve(context.Background(), exec.ExecutionID, "admin-operator")
	if err != nil {
		t.Fatalf("approval error: %v", err)
	}

	updated, ok := engine.GetExecution(exec.ExecutionID)
	if !ok {
		t.Fatalf("expected to find execution %s", exec.ExecutionID)
	}
	if updated.Status != StatusSucceeded {
		t.Errorf("expected status SUCCEEDED, got %s", updated.Status)
	}
	if updated.ApprovedBy != "admin-operator" {
		t.Errorf("expected approvedBy admin-operator, got %s", updated.ApprovedBy)
	}

	// Test rollback
	err = engine.Rollback(context.Background(), exec.ExecutionID)
	if err != nil {
		t.Fatalf("rollback error: %v", err)
	}
	rolledBack, _ := engine.GetExecution(exec.ExecutionID)
	if rolledBack.Status != StatusRolledBack {
		t.Errorf("expected status ROLLED_BACK, got %s", rolledBack.Status)
	}
}

func TestPlaybookEngine_AutoExecuteWithoutApproval(t *testing.T) {
	engine := NewPlaybookEngine(nil)

	// Trigger port flapping playbook (requires_approval = false)
	exec, err := engine.EvaluateAndTrigger(context.Background(), "port_flapping", 0.92, "access-sw-1")
	if err != nil {
		t.Fatalf("unexpected trigger error: %v", err)
	}
	if exec == nil {
		t.Fatalf("expected execution record, got nil")
	}

	if exec.Status != StatusSucceeded {
		t.Errorf("expected auto-execution status SUCCEEDED, got %s", exec.Status)
	}
}
