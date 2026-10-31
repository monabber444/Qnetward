package ctwatcher

import "time"

// CTCertificate represents a certificate observed in a Certificate Transparency log.
type CTCertificate struct {
	Domain       string    `json:"domain"`
	SANs         []string  `json:"sans"`
	Issuer       string    `json:"issuer"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	SerialNumber string    `json:"serial_number"`
	LogSource    string    `json:"log_source"`
}

// DetectionType categorizes how a lookalike domain was detected.
type DetectionType string

const (
	DetectLevenshtein            DetectionType = "levenshtein"
	DetectHomoglyph              DetectionType = "homoglyph"
	DetectSubdomainImpersonation DetectionType = "subdomain_impersonation"
	DetectKeywordCombination     DetectionType = "keyword_combination"
)

// TyposquatFinding details a detected potential phishing or lookalike domain.
type TyposquatFinding struct {
	ID              string        `json:"id"`
	MonitoredBrand  string        `json:"monitored_brand"`
	SuspectDomain   string        `json:"suspect_domain"`
	DetectionType   DetectionType `json:"detection_type"`
	SimilarityScore float64       `json:"similarity_score"` // 0.0 to 1.0 (1.0 = identical)
	Issuer          string        `json:"issuer"`
	DiscoveredAt    time.Time     `json:"discovered_at"`
	Severity        string        `json:"severity"` // "HIGH", "MEDIUM", "LOW"
	Status          string        `json:"status"`   // "INVESTIGATING", "CONFIRMED_PHISHING", "BENIGN"
}

// WatcherConfig configures the Certificate Transparency watcher.
type WatcherConfig struct {
	Brands            []string      `json:"brands"`             // Primary domain brands to protect
	SimilarityThreshold float64     `json:"similarity_threshold"` // Threshold to trigger alert (e.g. 0.75)
	CheckHomoglyphs   bool          `json:"check_homoglyphs"`
	CheckSubdomains   bool          `json:"check_subdomains"`
}
