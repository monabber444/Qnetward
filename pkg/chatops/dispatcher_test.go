package chatops

import (
	"context"
	"strings"
	"testing"

	"netwatcher/pkg/playbooks"
)

func TestChatOpsDispatcher_Commands(t *testing.T) {
	pbEngine := playbooks.NewPlaybookEngine(nil)
	dispatcher := NewDispatcher(pbEngine, "https://hooks.slack.com/services/test")

	// 1. Status command
	res := dispatcher.ExecuteCommand(context.Background(), "/qnetward status", "ops-user")
	if !res.Success || !strings.Contains(res.Message, "Operational") {
		t.Errorf("expected successful status response, got %+v", res)
	}

	// 2. Mute command
	res = dispatcher.ExecuteCommand(context.Background(), "/qnetward mute core-router-1 2h", "ops-user")
	if !res.Success || !strings.Contains(res.Message, "muted for `2h`") {
		t.Errorf("expected successful mute response, got %+v", res)
	}

	// 3. Trigger and Approve command via ChatOps
	exec, err := pbEngine.EvaluateAndTrigger(context.Background(), "optical_transceiver_degraded", 0.90, "core-router-1")
	if err != nil || exec == nil {
		t.Fatalf("failed to trigger playbook: %v", err)
	}

	res = dispatcher.ExecuteCommand(context.Background(), "/qnetward approve "+exec.ExecutionID, "alice")
	if !res.Success || !strings.Contains(res.Message, "approved and executed") {
		t.Errorf("expected successful approve response, got %+v", res)
	}

	// 4. Rollback command
	res = dispatcher.ExecuteCommand(context.Background(), "/qnetward rollback "+exec.ExecutionID, "alice")
	if !res.Success || !strings.Contains(res.Message, "successfully rolled back") {
		t.Errorf("expected successful rollback response, got %+v", res)
	}
}
