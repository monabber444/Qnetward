package securityengine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"netwatcher/pkg/netguard"
)

var (
	ErrOwnershipVerificationRequired = errors.New("403 Ownership Verification Required: active scanning requires verified asset ownership")
	ErrSSRFBlocked                   = errors.New("security violation: target IP address is in a prohibited or private range")
	ErrVerifiedBusinessRequired      = errors.New("403 Forbidden: Tier 2 deep authorized scans require a verified business plan")
	ErrScopeWindowExpired            = errors.New("403 Forbidden: current timestamp is outside authorized scanning window")
	ErrKillSwitchTriggered           = errors.New("scan aborted: emergency kill switch is active for target")
	ErrRateLimitExceeded             = errors.New("429 Too Many Requests: security scan rate limit token bucket depleted")
)

type ScanTier string

const (
	Tier0_Passive        ScanTier = "tier_0_passive"
	Tier1_SafeActive     ScanTier = "tier_1_safe_active"
	Tier2_DeepAuthorized ScanTier = "tier_2_deep_authorized"
)

type ScanAuthorizationRequest struct {
	ActorID          string    `json:"actor_id"`
	TenantID         int64     `json:"tenant_id"`
	Target           string    `json:"target"`
	Tier             ScanTier  `json:"tier"`
	Action           string    `json:"action"` // e.g. "port_scan", "tls_grade", "cve_match", "dast_crawler"
	VerifiedBusiness bool      `json:"verified_business"`
	ScopeStart       time.Time `json:"scope_start"`
	ScopeEnd         time.Time `json:"scope_end"`
	KillSwitchActive bool      `json:"kill_switch_active"`
}

type PolicyEngine struct {
	verifier    *OwnershipVerifier
	auditLogger AuditLogger
	netguardOpt netguard.Options
}

func NewPolicyEngine(verifier *OwnershipVerifier, auditLogger AuditLogger) *PolicyEngine {
	if auditLogger == nil {
		auditLogger = NewMemoryAuditLogger()
	}
	return &PolicyEngine{
		verifier:    verifier,
		auditLogger: auditLogger,
		netguardOpt: netguard.Options{
			AllowPrivate:  false,
			AllowLoopback: false,
		},
	}
}

// SetNetguardOptions allows configuring intranet or self-hosted SSRF allowances.
func (p *PolicyEngine) SetNetguardOptions(opts netguard.Options) {
	p.netguardOpt = opts
}

// AuthorizeScan evaluates the 3-Tier security policy, enforces SSRF defenses,
// checks asset ownership, and logs an immutable audit trail.
func (p *PolicyEngine) AuthorizeScan(ctx context.Context, req ScanAuthorizationRequest) error {
	now := time.Now().UTC()

	// 1. SSRF Defense-in-depth check
	if err := netguard.ValidateTarget(ctx, req.Target, p.netguardOpt); err != nil {
		_ = p.auditLogger.Log(ctx, AuditLogEntry{
			ActorID:   req.ActorID,
			TenantID:  req.TenantID,
			Target:    req.Target,
			Tier:      string(req.Tier),
			Action:    req.Action,
			Outcome:   OutcomeBlocked,
			Reason:    fmt.Sprintf("SSRF blocked: %v", err),
			Timestamp: now,
		})
		return fmt.Errorf("%w: %v", ErrSSRFBlocked, err)
	}

	// 2. Tier 0 (Passive): Allowed on any target
	if req.Tier == Tier0_Passive {
		_ = p.auditLogger.Log(ctx, AuditLogEntry{
			ActorID:   req.ActorID,
			TenantID:  req.TenantID,
			Target:    req.Target,
			Tier:      string(req.Tier),
			Action:    req.Action,
			Outcome:   OutcomeAllowed,
			Reason:    "Tier 0 passive inspection permitted",
			Timestamp: now,
		})
		return nil
	}

	// 3. Tier 1 (Safe Active): Requires verified ownership
	if !p.verifier.IsVerified(req.TenantID, req.Target) {
		_ = p.auditLogger.Log(ctx, AuditLogEntry{
			ActorID:   req.ActorID,
			TenantID:  req.TenantID,
			Target:    req.Target,
			Tier:      string(req.Tier),
			Action:    req.Action,
			Outcome:   OutcomeDenied,
			Reason:    "Unverified asset ownership",
			Timestamp: now,
		})
		return ErrOwnershipVerificationRequired
	}

	if req.Tier == Tier1_SafeActive {
		_ = p.auditLogger.Log(ctx, AuditLogEntry{
			ActorID:   req.ActorID,
			TenantID:  req.TenantID,
			Target:    req.Target,
			Tier:      string(req.Tier),
			Action:    req.Action,
			Outcome:   OutcomeAllowed,
			Reason:    "Tier 1 authorized: asset verified",
			Timestamp: now,
		})
		return nil
	}

	// 4. Tier 2 (Deep Authorized): Verified Business + Active Scope Window + Kill Switch
	if req.Tier == Tier2_DeepAuthorized {
		if req.KillSwitchActive {
			_ = p.auditLogger.Log(ctx, AuditLogEntry{
				ActorID:   req.ActorID,
				TenantID:  req.TenantID,
				Target:    req.Target,
				Tier:      string(req.Tier),
				Action:    req.Action,
				Outcome:   OutcomeDenied,
				Reason:    "Emergency kill switch active",
				Timestamp: now,
			})
			return ErrKillSwitchTriggered
		}

		if !req.VerifiedBusiness {
			_ = p.auditLogger.Log(ctx, AuditLogEntry{
				ActorID:   req.ActorID,
				TenantID:  req.TenantID,
				Target:    req.Target,
				Tier:      string(req.Tier),
				Action:    req.Action,
				Outcome:   OutcomeDenied,
				Reason:    "Verified business plan required for Tier 2",
				Timestamp: now,
			})
			return ErrVerifiedBusinessRequired
		}

		// Check scope window if configured
		if !req.ScopeStart.IsZero() && now.Before(req.ScopeStart) {
			_ = p.auditLogger.Log(ctx, AuditLogEntry{
				ActorID:   req.ActorID,
				TenantID:  req.TenantID,
				Target:    req.Target,
				Tier:      string(req.Tier),
				Action:    req.Action,
				Outcome:   OutcomeDenied,
				Reason:    "Scope window has not started yet",
				Timestamp: now,
			})
			return ErrScopeWindowExpired
		}
		if !req.ScopeEnd.IsZero() && now.After(req.ScopeEnd) {
			_ = p.auditLogger.Log(ctx, AuditLogEntry{
				ActorID:   req.ActorID,
				TenantID:  req.TenantID,
				Target:    req.Target,
				Tier:      string(req.Tier),
				Action:    req.Action,
				Outcome:   OutcomeDenied,
				Reason:    "Scope window has expired",
				Timestamp: now,
			})
			return ErrScopeWindowExpired
		}

		_ = p.auditLogger.Log(ctx, AuditLogEntry{
			ActorID:   req.ActorID,
			TenantID:  req.TenantID,
			Target:    req.Target,
			Tier:      string(req.Tier),
			Action:    req.Action,
			Outcome:   OutcomeAllowed,
			Reason:    "Tier 2 deep authorized scan approved within scope window",
			Timestamp: now,
		})
		return nil
	}

	return fmt.Errorf("unknown scan tier %s", req.Tier)
}
