package securityengine_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"netwatcher/pkg/securityengine"
)

func TestAssetOwnershipVerification(t *testing.T) {
	verifier := securityengine.NewOwnershipVerifier()
	ctx := context.Background()

	// 1. DNS TXT Verification
	pendingDNS := verifier.RegisterPending(10, "example.com", securityengine.MethodDNSTXT)
	if pendingDNS.Status != securityengine.StatusPending {
		t.Fatalf("expected pending status, got %s", pendingDNS.Status)
	}

	// Mock DNS query returning the expected TXT token
	verifier.SetLookupHooks(
		func(c context.Context, name string) ([]string, error) {
			if name == "example.com" {
				return []string{"v=spf1 include:_spf.google.com ~all", "qnetward-verify=" + pendingDNS.Token}, nil
			}
			return nil, errors.New("nxdomain")
		},
		nil,
	)

	if err := verifier.Verify(ctx, 10, "example.com"); err != nil {
		t.Fatalf("DNS TXT verification failed: %v", err)
	}
	if !verifier.IsVerified(10, "example.com") {
		t.Fatalf("expected example.com to be verified")
	}

	// 2. HTTP Well-Known Verification
	pendingHTTP := verifier.RegisterPending(10, "mysite.org", securityengine.MethodHTTPWellKnown)
	verifier.SetLookupHooks(
		nil,
		func(c context.Context, url string) (string, error) {
			if strings.Contains(url, "mysite.org/.well-known/qnetward-verify.txt") {
				return pendingHTTP.Token, nil
			}
			return "", errors.New("404 not found")
		},
	)

	if err := verifier.Verify(ctx, 10, "mysite.org"); err != nil {
		t.Fatalf("HTTP well-known verification failed: %v", err)
	}
	if !verifier.IsVerified(10, "mysite.org") {
		t.Fatalf("expected mysite.org to be verified")
	}
}

func TestSecurityPolicyThreeTierEnforcement(t *testing.T) {
	verifier := securityengine.NewOwnershipVerifier()
	audit := securityengine.NewMemoryAuditLogger()
	engine := securityengine.NewPolicyEngine(verifier, audit)
	ctx := context.Background()

	// Verify one target
	verifiedTarget := "93.184.216.34"
	unverifiedTarget := "8.8.8.8"

	_ = verifier.RegisterPending(100, verifiedTarget, securityengine.MethodDNSTXT)
	verifier.SetLookupHooks(
		func(c context.Context, name string) ([]string, error) {
			rec, _ := verifier.GetVerification(100, verifiedTarget)
			return []string{"qnetward-verify=" + rec.Token}, nil
		},
		nil,
	)
	_ = verifier.Verify(ctx, 100, verifiedTarget)

	// 1. Tier 0 (Passive): Allowed on ANY target without ownership verification
	err := engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:  "user-1",
		TenantID: 100,
		Target:   unverifiedTarget,
		Tier:     securityengine.Tier0_Passive,
		Action:   "ct_log_monitor",
	})
	if err != nil {
		t.Fatalf("Tier 0 passive should be allowed on unverified target, got: %v", err)
	}

	// 2. Tier 1 (Safe Active): Strictly denied on unverified target
	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:  "user-1",
		TenantID: 100,
		Target:   unverifiedTarget,
		Tier:     securityengine.Tier1_SafeActive,
		Action:   "port_scan",
	})
	if !errors.Is(err, securityengine.ErrOwnershipVerificationRequired) {
		t.Fatalf("expected ErrOwnershipVerificationRequired, got: %v", err)
	}

	// Tier 1 on verified target: Allowed!
	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:  "user-1",
		TenantID: 100,
		Target:   verifiedTarget,
		Tier:     securityengine.Tier1_SafeActive,
		Action:   "port_scan",
	})
	if err != nil {
		t.Fatalf("Tier 1 should be allowed on verified target, got: %v", err)
	}

	// 3. SSRF Target Blocking (169.254.169.254 cloud metadata & loopback)
	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:  "user-attacker",
		TenantID: 100,
		Target:   "169.254.169.254",
		Tier:     securityengine.Tier1_SafeActive,
		Action:   "port_scan",
	})
	if !errors.Is(err, securityengine.ErrSSRFBlocked) {
		t.Fatalf("expected ErrSSRFBlocked for metadata target, got: %v", err)
	}

	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:  "user-attacker",
		TenantID: 100,
		Target:   "127.0.0.1",
		Tier:     securityengine.Tier0_Passive,
		Action:   "whois",
	})
	if !errors.Is(err, securityengine.ErrSSRFBlocked) {
		t.Fatalf("expected ErrSSRFBlocked for loopback target, got: %v", err)
	}

	// 4. Tier 2 (Deep Authorized)
	// Without VerifiedBusiness flag -> rejected
	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:          "user-1",
		TenantID:         100,
		Target:           verifiedTarget,
		Tier:             securityengine.Tier2_DeepAuthorized,
		Action:           "dast_crawler",
		VerifiedBusiness: false,
	})
	if !errors.Is(err, securityengine.ErrVerifiedBusinessRequired) {
		t.Fatalf("expected ErrVerifiedBusinessRequired, got: %v", err)
	}

	// With KillSwitchActive -> rejected
	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:          "user-1",
		TenantID:         100,
		Target:           verifiedTarget,
		Tier:             securityengine.Tier2_DeepAuthorized,
		Action:           "dast_crawler",
		VerifiedBusiness: true,
		KillSwitchActive: true,
	})
	if !errors.Is(err, securityengine.ErrKillSwitchTriggered) {
		t.Fatalf("expected ErrKillSwitchTriggered, got: %v", err)
	}

	// Scope window expired -> rejected
	pastTime := time.Now().Add(-2 * time.Hour)
	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:          "user-1",
		TenantID:         100,
		Target:           verifiedTarget,
		Tier:             securityengine.Tier2_DeepAuthorized,
		Action:           "dast_crawler",
		VerifiedBusiness: true,
		ScopeEnd:         pastTime,
	})
	if !errors.Is(err, securityengine.ErrScopeWindowExpired) {
		t.Fatalf("expected ErrScopeWindowExpired, got: %v", err)
	}

	// Valid Tier 2 request -> Approved!
	futureTime := time.Now().Add(2 * time.Hour)
	err = engine.AuthorizeScan(ctx, securityengine.ScanAuthorizationRequest{
		ActorID:          "user-1",
		TenantID:         100,
		Target:           verifiedTarget,
		Tier:             securityengine.Tier2_DeepAuthorized,
		Action:           "dast_crawler",
		VerifiedBusiness: true,
		ScopeStart:       time.Now().Add(-10 * time.Minute),
		ScopeEnd:         futureTime,
	})
	if err != nil {
		t.Fatalf("valid Tier 2 should be approved, got: %v", err)
	}

	// 5. Verify Immutable Audit Logs
	entries := audit.GetEntries(100)
	if len(entries) < 7 {
		t.Fatalf("expected at least 7 audit log entries, got %d", len(entries))
	}
}

func TestCVECatalogAndPassiveAudit(t *testing.T) {
	// 1. CVE Catalog
	catalog := securityengine.NewCVECatalog()
	matches := catalog.MatchService("nginx", "1.20")
	if len(matches) == 0 {
		t.Fatalf("expected CVE matches for nginx")
	}
	if matches[0].CVEID != "CVE-2023-44487" {
		t.Fatalf("unexpected CVE ID: %s", matches[0].CVEID)
	}
	if !matches[0].InCISAKEV {
		t.Fatalf("expected CVE-2023-44487 to be in CISA KEV")
	}

	// 2. Mail Security Passive Audit
	spf := "v=spf1 include:_spf.google.com ~all"
	dmarc := "v=DMARC1; p=reject; rua=mailto:dmarc@example.com"
	mailAudit := securityengine.AuditMailSecurity("example.com", spf, dmarc)
	if !mailAudit.HasSPF || !mailAudit.HasDMARC {
		t.Fatalf("expected valid SPF and DMARC: %+v", mailAudit)
	}
	if !mailAudit.DMARCCompliant || mailAudit.DMARCPolicy != "reject" {
		t.Fatalf("expected compliant DMARC reject policy")
	}

	// 3. Certificate Transparency Passive Audit
	ct := securityengine.AuditCertificateTransparency("example.com")
	if ct.CertCount <= 0 || !ct.WildcardSeen {
		t.Fatalf("unexpected CT audit: %+v", ct)
	}
}
