# 16. Autonomous Self-Healing Playbooks, Closed-Loop Remediation, and ChatOps

## Status

Accepted

## Context

Observability platforms conventionally stop at passive alerting, requiring human on-call engineers to wake up, diagnose cascading alert storms, and execute manual CLI commands (e.g. BGP path prepending, switch port shutdown, cloud firewall quarantine).

This manual workflow causes extended Mean Time to Resolution (MTTR), human error under outage pressure, and engineer burnout.

To achieve autonomous self-healing networks, Qnetward requires:
1. **Declarative Auto-Remediation Playbooks**: Deterministic workflows triggered directly by AI Root-Cause Analysis (RCA) or metric threshold breaches.
2. **Interactive Operator Approval Gates & Safe Rollback**: For critical production traffic, the platform must support an approval state (`PENDING_APPROVAL`) where operators can review the action, simulated blast radius, and execute with a single click, or roll back if health verification fails.
3. **Bidirectional ChatOps**: Enabling operators to approve playbooks, silence device alerts, and inspect RCA directly from Slack, Microsoft Teams, or Discord.

## Decision

1. **Autonomous Playbook Engine (`pkg/playbooks`)**:
   - **`PlaybookEngine` (`pkg/playbooks/engine.go`)**: Evaluates incoming incident triggers (`optical_transceiver_degraded`, `port_flapping`, `cascading_outage`) against registered playbook definitions.
   - **Supported Remediation Actions**:
     - `reroute_bgp`: Shifts transit traffic away from degrading links via AS-path prepending or community tagging.
     - `quarantine_port`: Administratively isolates flapping access switch ports to prevent spanning-tree storms.
     - `cloud_sg_quarantine`: Injects temporary drop rules into cloud firewalls (AWS SG, Azure NSG, GCP Firewall).
     - `failover_circuit`: Triggers backup WAN / secondary ISP failover.
     - `dispatch_ticket`: Automatically generates P1/P2 tickets in Jira Service Management, ServiceNow, or PagerDuty.
   - **Approval & Rollback Controls**: High-impact playbooks require explicit operator approval (`Approve`). If health conditions do not normalize, the engine executes an automated `Rollback` restoring baseline configuration.

2. **Bidirectional ChatOps & Webhooks (`pkg/chatops`)**:
   - **`Dispatcher` (`pkg/chatops/dispatcher.go`)**: Formats rich Slack / MS Teams / Webhook alert cards with root-cause summaries, confidence scores, and action links.
   - **Slash Command Parser**:
     - `/qnetward status`: Instant cluster health and probe node count.
     - `/qnetward approve <exec_id>`: Approves pending remediation playbooks directly from chat.
     - `/qnetward rollback <exec_id>`: Rolls back executed actions to baseline.
     - `/qnetward mute <target> <duration>`: Silences device alerts for maintenance windows.

3. **Web Cockpit AIOps Workspace (`/aiops`)**:
   - Dedicated workspace (`src/features/aiops/aiops-workspace.tsx` & `src/app/(protected)/aiops/page.tsx`).
   - Remediation Execution History with live interactive **`Approve & Execute`** and **`Rollback`** buttons.
   - Real-time terminal log viewer displaying action step outputs.
   - ChatOps slash command reference and active webhook status indicators.

## Consequences

- Cuts MTTR from hours to seconds for common network failure modes.
- Prevents cascading alert storms from propagating across distribution layers.
- Operators maintain complete control via zero-friction approval gates in the web dashboard or mobile chat.
