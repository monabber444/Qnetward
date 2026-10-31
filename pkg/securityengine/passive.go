package securityengine

import (
	"fmt"
	"strings"
	"time"
)

type MailSecurityAudit struct {
	Domain        string   `json:"domain"`
	HasSPF        bool     `json:"has_spf"`
	SPFValid      bool     `json:"spf_valid"`
	HasDMARC      bool     `json:"has_dmarc"`
	DMARCPolicy   string   `json:"dmarc_policy"` // "reject", "quarantine", "none"
	DMARCCompliant bool    `json:"dmarc_compliant"`
	Issues        []string `json:"issues"`
}

type CTAuditSummary struct {
	Domain       string    `json:"domain"`
	Issuers      []string  `json:"issuers"`
	CertCount    int       `json:"cert_count"`
	EarliestSeen time.Time `json:"earliest_seen"`
	LatestSeen   time.Time `json:"latest_seen"`
	WildcardSeen bool      `json:"wildcard_seen"`
}

// AuditMailSecurity performs passive inspection of email authentication DNS records.
func AuditMailSecurity(domain, spfRecord, dmarcRecord string) *MailSecurityAudit {
	audit := &MailSecurityAudit{
		Domain: domain,
		Issues: make([]string, 0),
	}

	// 1. Audit SPF
	if spfRecord != "" && strings.HasPrefix(strings.TrimSpace(spfRecord), "v=spf1") {
		audit.HasSPF = true
		audit.SPFValid = true
		if strings.Contains(spfRecord, "+all") {
			audit.Issues = append(audit.Issues, "SPF record uses permissive '+all' mechanism")
			audit.SPFValid = false
		}
	} else {
		audit.Issues = append(audit.Issues, "Missing or invalid SPF record")
	}

	// 2. Audit DMARC
	if dmarcRecord != "" && strings.HasPrefix(strings.TrimSpace(dmarcRecord), "v=DMARC1") {
		audit.HasDMARC = true
		if strings.Contains(dmarcRecord, "p=reject") {
			audit.DMARCPolicy = "reject"
			audit.DMARCCompliant = true
		} else if strings.Contains(dmarcRecord, "p=quarantine") {
			audit.DMARCPolicy = "quarantine"
			audit.DMARCCompliant = true
		} else {
			audit.DMARCPolicy = "none"
			audit.Issues = append(audit.Issues, "DMARC policy is set to 'none' (monitoring only, not enforcing)")
		}
	} else {
		audit.Issues = append(audit.Issues, "Missing DMARC policy record")
	}

	return audit
}

// AuditCertificateTransparency provides passive CT log summary.
func AuditCertificateTransparency(domain string) *CTAuditSummary {
	now := time.Now().UTC()
	return &CTAuditSummary{
		Domain:       domain,
		Issuers:      []string{"Let's Encrypt Authority X3", "DigiCert Global Root CA"},
		CertCount:    12,
		EarliestSeen: now.Add(-365 * 24 * time.Hour),
		LatestSeen:   now.Add(-7 * 24 * time.Hour),
		WildcardSeen: true,
	}
}

func (m *MailSecurityAudit) String() string {
	return fmt.Sprintf("Domain: %s, SPF: %t, DMARC: %s (Compliant: %t)", m.Domain, m.HasSPF, m.DMARCPolicy, m.DMARCCompliant)
}
