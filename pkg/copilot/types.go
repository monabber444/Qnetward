package copilot

import "time"

// ArtifactType categorizes synthetic configurations or documentation produced by the Copilot.
type ArtifactType string

const (
	ArtifactCiscoACL        ArtifactType = "CISCO_ACL"
	ArtifactJuniperFilter   ArtifactType = "JUNIPER_FILTER"
	ArtifactLinuxNftables   ArtifactType = "NFTABLES_RULESET"
	ArtifactPostMortem      ArtifactType = "POSTMORTEM_REPORT"
	ArtifactRemediationPlan ArtifactType = "REMEDIATION_PLAN"
)

// CopilotQuery represents an incoming operator prompt and relevant network context.
type CopilotQuery struct {
	QueryID           string `json:"query_id"`
	SessionID         string `json:"session_id"`
	UserPrompt        string `json:"user_prompt"`
	TargetDevice      string `json:"target_device,omitempty"`
	ActiveIncidentID  string `json:"active_incident_id,omitempty"`
}

// GeneratedArtifact represents a deployable ACL, policy, or markdown post-mortem.
type GeneratedArtifact struct {
	Type        ArtifactType `json:"type"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	SyntaxLang  string       `json:"syntax_lang"` // "cisco", "junos", "markdown"
}

// CopilotResponse contains synthesized guidance, actionable next steps, and artifacts.
type CopilotResponse struct {
	QueryID          string               `json:"query_id"`
	Explanation      string               `json:"explanation"`
	ConfidenceScore  float64              `json:"confidence_score"` // 0.0 to 1.0
	Artifact         *GeneratedArtifact   `json:"artifact,omitempty"`
	SuggestedActions []string             `json:"suggested_actions"`
	GeneratedAt      time.Time            `json:"generated_at"`
}
