package dast

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
)

const (
	TopicDASTFinding = "dast.finding"
)

// Auditor conducts automated defensive scans against HTTP/HTTPS endpoints.
type Auditor struct {
	client *http.Client
	bus    eventbus.Bus
	mu     sync.RWMutex
	reports []*AuditReport
}

// NewAuditor creates a defensive HTTP security auditor.
func NewAuditor(client *http.Client, bus eventbus.Bus) *Auditor {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
		}
	}
	return &Auditor{
		client:  client,
		bus:     bus,
		reports: make([]*AuditReport, 0),
	}
}

// AuditTarget inspects an endpoint and generates security findings.
func (a *Auditor) AuditTarget(ctx context.Context, targetURL string, cfg AuditConfig) (*AuditReport, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit request: %w", err)
	}

	// Add Origin header to test CORS posture
	req.Header.Set("Origin", "https://untrusted-cross-origin.test")
	req.Header.Set("User-Agent", "Qnetward-CASM-Auditor/1.0 (+https://github.com/Quantum-Axis/Qnetward)")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()
	var findings []*DASTFinding

	// 1. Audit HTTP Security Headers
	if cfg.CheckHeaders {
		findings = append(findings, a.auditHeaders(targetURL, resp.Header)...)
	}

	// 2. Audit CORS configuration
	if cfg.CheckCORS {
		findings = append(findings, a.auditCORS(targetURL, resp.Header)...)
	}

	// 3. Audit Cookies
	if cfg.CheckCookies {
		findings = append(findings, a.auditCookies(targetURL, resp.Cookies(), strings.HasPrefix(targetURL, "https://"))...)
	}

	// 4. Audit Information Leakage
	if cfg.CheckInfoLeakage {
		findings = append(findings, a.auditInfoLeakage(targetURL, resp.Header)...)
	}

	grade := calculateGrade(findings)

	report := &AuditReport{
		TargetURL:     targetURL,
		DurationMs:    duration,
		StatusCode:    resp.StatusCode,
		Findings:      findings,
		AuditedAt:     time.Now(),
		SecurityGrade: grade,
	}

	a.mu.Lock()
	a.reports = append(a.reports, report)
	a.mu.Unlock()

	// Publish each finding to eventbus
	if a.bus != nil {
		for _, f := range findings {
			if payload, err := json.Marshal(f); err == nil {
				_ = a.bus.Publish(ctx, TopicDASTFinding, payload)
			}
		}
	}

	return report, nil
}

func (a *Auditor) auditHeaders(target string, h http.Header) []*DASTFinding {
	var findings []*DASTFinding
	isHTTPS := strings.HasPrefix(target, "https://")

	if isHTTPS && h.Get("Strict-Transport-Security") == "" {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("hsts-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "HEADER_MISSING",
			Title:       "Missing HTTP Strict Transport Security (HSTS)",
			Description: "The endpoint does not send Strict-Transport-Security header, allowing potential SSL-stripping MITM attacks.",
			Severity:    SeverityHigh,
			Remediation: "Add 'Strict-Transport-Security: max-age=31536000; includeSubDomains; preload' to all HTTPS responses.",
			DetectedAt:  time.Now(),
		})
	}

	if h.Get("Content-Security-Policy") == "" {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("csp-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "HEADER_MISSING",
			Title:       "Missing Content Security Policy (CSP)",
			Description: "No Content-Security-Policy header detected. The application is at higher risk of cross-site scripting (XSS) and data injection.",
			Severity:    SeverityMedium,
			Remediation: "Configure a strict Content-Security-Policy header restricting trusted script sources.",
			DetectedAt:  time.Now(),
		})
	}

	if h.Get("X-Frame-Options") == "" && !strings.Contains(h.Get("Content-Security-Policy"), "frame-ancestors") {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("xfo-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "HEADER_MISSING",
			Title:       "Missing Anti-Clickjacking Protection",
			Description: "Neither X-Frame-Options nor frame-ancestors CSP directive is configured, leaving the endpoint susceptible to UI clickjacking.",
			Severity:    SeverityMedium,
			Remediation: "Set 'X-Frame-Options: DENY' or 'X-Frame-Options: SAMEORIGIN'.",
			DetectedAt:  time.Now(),
		})
	}

	if strings.ToLower(h.Get("X-Content-Type-Options")) != "nosniff" {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("xcto-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "HEADER_MISSING",
			Title:       "Missing X-Content-Type-Options: nosniff",
			Description: "The browser may perform MIME sniffing on untrusted response bodies.",
			Severity:    SeverityLow,
			Remediation: "Include 'X-Content-Type-Options: nosniff' header.",
			DetectedAt:  time.Now(),
		})
	}

	return findings
}

func (a *Auditor) auditCORS(target string, h http.Header) []*DASTFinding {
	var findings []*DASTFinding
	acao := h.Get("Access-Control-Allow-Origin")
	acac := strings.ToLower(h.Get("Access-Control-Allow-Credentials"))

	if acao == "*" && acac == "true" {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("cors-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "CORS_PERMISSIVE",
			Title:       "Wildcard CORS Origin With Credentials Allowed",
			Description: "Access-Control-Allow-Origin is configured as wildcard while Access-Control-Allow-Credentials is true, exposing authenticated user data across origins.",
			Severity:    SeverityHigh,
			Remediation: "Do not use wildcard '*' with credentials. Specify explicit trusted origin domains.",
			Evidence:    fmt.Sprintf("ACAO: %s, ACAC: %s", acao, acac),
			DetectedAt:  time.Now(),
		})
	} else if strings.Contains(acao, "untrusted-cross-origin.test") {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("cors-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "CORS_PERMISSIVE",
			Title:       "Arbitrary Origin Reflection in CORS",
			Description: "The endpoint blindly echoes untrusted Origin headers into Access-Control-Allow-Origin.",
			Severity:    SeverityHigh,
			Remediation: "Validate Origin headers against an explicit, strict whitelist of trusted hosts.",
			Evidence:    fmt.Sprintf("Reflected Origin: %s", acao),
			DetectedAt:  time.Now(),
		})
	}

	return findings
}

func (a *Auditor) auditCookies(target string, cookies []*http.Cookie, isHTTPS bool) []*DASTFinding {
	var findings []*DASTFinding

	for _, c := range cookies {
		if isHTTPS && !c.Secure {
			findings = append(findings, &DASTFinding{
				ID:          fmt.Sprintf("cookie-sec-%d", time.Now().UnixNano()%1000000),
				TargetURL:   target,
				Category:    "COOKIE_INSECURE",
				Title:       fmt.Sprintf("Cookie '%s' Missing Secure Flag", c.Name),
				Description: "Cookie transmitted over HTTPS without the Secure flag can be leaked over plaintext HTTP connections.",
				Severity:    SeverityMedium,
				Remediation: "Set the 'Secure' attribute on all cookies over HTTPS.",
				DetectedAt:  time.Now(),
			})
		}
		if !c.HttpOnly {
			findings = append(findings, &DASTFinding{
				ID:          fmt.Sprintf("cookie-ho-%d", time.Now().UnixNano()%1000000),
				TargetURL:   target,
				Category:    "COOKIE_INSECURE",
				Title:       fmt.Sprintf("Cookie '%s' Missing HttpOnly Flag", c.Name),
				Description: "Cookie accessible via document.cookie JavaScript API, increasing session hijacking impact from XSS.",
				Severity:    SeverityLow,
				Remediation: "Set the 'HttpOnly' flag for sensitive cookies.",
				DetectedAt:  time.Now(),
			})
		}
	}

	return findings
}

func (a *Auditor) auditInfoLeakage(target string, h http.Header) []*DASTFinding {
	var findings []*DASTFinding

	server := h.Get("Server")
	if server != "" && (strings.Contains(server, "/") || strings.ContainsAny(server, "0123456789")) {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("leak-server-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "INFO_LEAKAGE",
			Title:       "Detailed Server Version Banner Disclosed",
			Description: fmt.Sprintf("Server header reveals specific software version (%s), assisting attackers with CVE targeting.", server),
			Severity:    SeverityLow,
			Remediation: "Configure the web server or reverse proxy to omit detailed version banners.",
			Evidence:    server,
			DetectedAt:  time.Now(),
		})
	}

	xPowered := h.Get("X-Powered-By")
	if xPowered != "" {
		findings = append(findings, &DASTFinding{
			ID:          fmt.Sprintf("leak-powered-%d", time.Now().UnixNano()%1000000),
			TargetURL:   target,
			Category:    "INFO_LEAKAGE",
			Title:       "Technology Stack Disclosed via X-Powered-By",
			Description: fmt.Sprintf("Response discloses backend technology: %s", xPowered),
			Severity:    SeverityLow,
			Remediation: "Disable X-Powered-By header in application framework settings.",
			Evidence:    xPowered,
			DetectedAt:  time.Now(),
		})
	}

	return findings
}

func calculateGrade(findings []*DASTFinding) string {
	hasCritical := false
	hasHigh := false
	mediumCount := 0

	for _, f := range findings {
		switch f.Severity {
		case SeverityCritical:
			hasCritical = true
		case SeverityHigh:
			hasHigh = true
		case SeverityMedium:
			mediumCount++
		}
	}

	if hasCritical {
		return "F"
	}
	if hasHigh {
		return "D"
	}
	if mediumCount >= 3 {
		return "C"
	}
	if mediumCount > 0 {
		return "B"
	}
	if len(findings) > 0 {
		return "A"
	}
	return "A+"
}

// GetReports returns all completed audit reports.
func (a *Auditor) GetReports() []*AuditReport {
	a.mu.RLock()
	defer a.mu.RUnlock()
	res := make([]*AuditReport, len(a.reports))
	copy(res, a.reports)
	return res
}
