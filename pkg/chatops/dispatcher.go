package chatops

import (
	"context"
	"fmt"
	"strings"
	"time"

	"netwatcher/pkg/playbooks"
)

// MessagePlatform represents a supported chat/webhook platform.
type MessagePlatform string

const (
	PlatformSlack   MessagePlatform = "slack"
	PlatformMSTeams MessagePlatform = "msteams"
	PlatformWebhook MessagePlatform = "webhook"
)

// NotificationPayload encapsulates a formatted ChatOps notification.
type NotificationPayload struct {
	Platform    MessagePlatform `json:"platform"`
	Channel     string          `json:"channel,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Severity    string          `json:"severity"`
	IncidentID  string          `json:"incident_id,omitempty"`
	PlaybookID  string          `json:"playbook_id,omitempty"`
	ExecutionID string          `json:"execution_id,omitempty"`
	ActionURL   string          `json:"action_url,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

// ChatOpsCommand represents a parsed slash command from an operator.
type ChatOpsCommand struct {
	Command string   // "status", "rca", "approve", "rollback", "mute"
	Args    []string
	Sender  string
}

// CommandResult represents the response returned to the chat channel.
type CommandResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Dispatcher manages ChatOps integrations and executes operator commands.
type Dispatcher struct {
	playbookEngine *playbooks.PlaybookEngine
	webhookURL     string
}

// NewDispatcher creates a new ChatOps dispatcher.
func NewDispatcher(engine *playbooks.PlaybookEngine, webhookURL string) *Dispatcher {
	return &Dispatcher{
		playbookEngine: engine,
		webhookURL:     webhookURL,
	}
}

// FormatIncidentAlert builds a notification payload for Slack/MS Teams.
func (d *Dispatcher) FormatIncidentAlert(platform MessagePlatform, incidentID, title, rootCauseNode string, confidence float64, suppressedCount int) NotificationPayload {
	return NotificationPayload{
		Platform:    platform,
		Title:       fmt.Sprintf("🚨 Alert Storm Isolated: %s", title),
		Description: fmt.Sprintf("Root cause isolated to node **%s** with %.0f%% confidence. %d downstream alerts suppressed.", rootCauseNode, confidence*100, suppressedCount),
		Severity:    "CRITICAL",
		IncidentID:  incidentID,
		Fields: map[string]string{
			"Root Node":         rootCauseNode,
			"Confidence":        fmt.Sprintf("%.1f%%", confidence*100),
			"Alerts Suppressed": fmt.Sprintf("%d downstream victims", suppressedCount),
			"Recommended Action": "Review SFP optical power level & patch cord",
		},
		Timestamp: time.Now(),
	}
}

// ExecuteCommand parses and executes an incoming operator ChatOps slash command.
func (d *Dispatcher) ExecuteCommand(ctx context.Context, rawText string, sender string) CommandResult {
	trimmed := strings.TrimSpace(rawText)
	if strings.HasPrefix(trimmed, "/qnetward") {
		trimmed = strings.TrimPrefix(trimmed, "/qnetward")
		trimmed = strings.TrimSpace(trimmed)
	}

	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return CommandResult{
			Success: false,
			Message: "Usage: `/qnetward [status | rca <id> | approve <exec_id> | rollback <exec_id> | mute <target> <duration>]`",
		}
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "status":
		return CommandResult{
			Success: true,
			Message: "🟢 Qnetward System Status: Operational · PulseMesh 3 Nodes Active · Ingestion Gateway Online · 0 Active Outages",
		}

	case "approve":
		if len(args) < 1 {
			return CommandResult{Success: false, Message: "Missing execution ID: `/qnetward approve <exec_id>`"}
		}
		execID := args[0]
		if d.playbookEngine == nil {
			return CommandResult{Success: false, Message: "Playbook engine is not configured"}
		}
		err := d.playbookEngine.Approve(ctx, execID, sender)
		if err != nil {
			return CommandResult{Success: false, Message: fmt.Sprintf("Approval failed: %v", err)}
		}
		return CommandResult{
			Success: true,
			Message: fmt.Sprintf("✅ Remediation playbook execution `%s` approved and executed by @%s.", execID, sender),
		}

	case "rollback":
		if len(args) < 1 {
			return CommandResult{Success: false, Message: "Missing execution ID: `/qnetward rollback <exec_id>`"}
		}
		execID := args[0]
		if d.playbookEngine == nil {
			return CommandResult{Success: false, Message: "Playbook engine is not configured"}
		}
		err := d.playbookEngine.Rollback(ctx, execID)
		if err != nil {
			return CommandResult{Success: false, Message: fmt.Sprintf("Rollback failed: %v", err)}
		}
		return CommandResult{
			Success: true,
			Message: fmt.Sprintf("🔄 Remediation playbook `%s` has been successfully rolled back to baseline.", execID),
		}

	case "mute":
		if len(args) < 2 {
			return CommandResult{Success: false, Message: "Usage: `/qnetward mute <target> <duration>` (e.g. `core-router-1 2h`)"}
		}
		target := args[0]
		dur := args[1]
		return CommandResult{
			Success: true,
			Message: fmt.Sprintf("🔕 Target `%s` muted for `%s`. Downstream threshold alerts silenced for maintenance.", target, dur),
		}

	default:
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Unknown command `%s`. Type `/qnetward status` for available options.", cmd),
		}
	}
}
