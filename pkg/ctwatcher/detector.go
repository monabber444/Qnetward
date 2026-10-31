package ctwatcher

import (
	"fmt"
	"strings"
	"time"
)

// BrandDetector evaluates domains for lookalike/phishing similarity.
type BrandDetector struct {
	config WatcherConfig
}

// NewBrandDetector creates a new brand impersonation detector.
func NewBrandDetector(cfg WatcherConfig) *BrandDetector {
	if cfg.SimilarityThreshold <= 0 {
		cfg.SimilarityThreshold = 0.75
	}
	return &BrandDetector{config: cfg}
}

// EvaluateDomain compares a discovered domain against monitored brands.
func (d *BrandDetector) EvaluateDomain(suspect string, cert CTCertificate) *TyposquatFinding {
	suspectClean := strings.ToLower(strings.TrimPrefix(suspect, "*."))

	for _, brand := range d.config.Brands {
		brandClean := strings.ToLower(strings.TrimSpace(brand))
		if suspectClean == brandClean {
			continue // Identical domain belongs to the brand owner
		}

		// Extract base name without TLD for typo analysis (e.g. "qnetward" from "qnetward.com")
		brandStem := extractStem(brandClean)
		suspectStem := extractStem(suspectClean)

		// 1. Subdomain impersonation check (e.g. "brand.evil-site.com" or "brand-login.com")
		if d.config.CheckSubdomains {
			if strings.HasPrefix(suspectClean, brandStem+".") ||
				strings.Contains(suspectClean, "."+brandStem+".") ||
				strings.HasPrefix(suspectClean, brandStem+"-") ||
				strings.Contains(suspectClean, "-"+brandStem) {
				return &TyposquatFinding{
					ID:              fmt.Sprintf("sq-%d", time.Now().UnixNano()%1000000),
					MonitoredBrand:  brandClean,
					SuspectDomain:   suspectClean,
					DetectionType:   DetectSubdomainImpersonation,
					SimilarityScore: 0.95,
					Issuer:          cert.Issuer,
					DiscoveredAt:    time.Now(),
					Severity:        "HIGH",
					Status:          "INVESTIGATING",
				}
			}
		}

		// 2. Homoglyph confusable check (e.g. "vv" -> "w", "0" -> "o")
		if d.config.CheckHomoglyphs {
			normalizedSuspect := normalizeConfusables(suspectStem)
			if normalizedSuspect == brandStem && suspectStem != brandStem {
				return &TyposquatFinding{
					ID:              fmt.Sprintf("sq-%d", time.Now().UnixNano()%1000000),
					MonitoredBrand:  brandClean,
					SuspectDomain:   suspectClean,
					DetectionType:   DetectHomoglyph,
					SimilarityScore: 0.98,
					Issuer:          cert.Issuer,
					DiscoveredAt:    time.Now(),
					Severity:        "HIGH",
					Status:          "INVESTIGATING",
				}
			}
		}

		// 3. Levenshtein edit distance check
		dist := LevenshteinDistance(brandStem, suspectStem)
		maxLen := len(brandStem)
		if len(suspectStem) > maxLen {
			maxLen = len(suspectStem)
		}
		if maxLen == 0 {
			continue
		}

		score := 1.0 - (float64(dist) / float64(maxLen))
		if score >= d.config.SimilarityThreshold && dist > 0 && dist <= 2 {
			severity := "MEDIUM"
			if score >= 0.85 {
				severity = "HIGH"
			}
			return &TyposquatFinding{
				ID:              fmt.Sprintf("sq-%d", time.Now().UnixNano()%1000000),
				MonitoredBrand:  brandClean,
				SuspectDomain:   suspectClean,
				DetectionType:   DetectLevenshtein,
				SimilarityScore: score,
				Issuer:          cert.Issuer,
				DiscoveredAt:    time.Now(),
				Severity:        severity,
				Status:          "INVESTIGATING",
			}
		}
	}

	return nil
}

// LevenshteinDistance calculates the minimum edit operations between two strings.
func LevenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	v0 := make([]int, lb+1)
	v1 := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		v0[j] = j
	}

	for i := 0; i < la; i++ {
		v1[0] = i + 1
		for j := 0; j < lb; j++ {
			cost := 0
			if a[i] != b[j] {
				cost = 1
			}
			min := v0[j+1] + 1 // deletion
			if v1[j]+1 < min { // insertion
				min = v1[j] + 1
			}
			if v0[j]+cost < min { // substitution
				min = v0[j] + cost
			}
			v1[j+1] = min
		}
		copy(v0, v1)
	}

	return v1[lb]
}

func normalizeConfusables(s string) string {
	r := s
	r = strings.ReplaceAll(r, "vv", "w")
	r = strings.ReplaceAll(r, "rn", "m")
	r = strings.ReplaceAll(r, "0", "o")
	r = strings.ReplaceAll(r, "1", "l")
	r = strings.ReplaceAll(r, "5", "s")
	return r
}

func extractStem(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return parts[0]
	}
	return domain
}
