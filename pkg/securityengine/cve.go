package securityengine

import (
	"strings"
	"sync"
)

// CVERecord combines NVD CVSS scoring, FIRST EPSS exploit probability, and CISA Known Exploited Vulnerabilities (KEV).
type CVERecord struct {
	CVEID        string   `json:"cve_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	CVSSScore    float64  `json:"cvss_score"`
	EPSSScore    float64  `json:"epss_score"` // 0.00 to 1.00 probability of in-the-wild exploitation
	InCISAKEV    bool     `json:"in_cisa_kev"`
	AffectedCPEs []string `json:"affected_cpes"`
}

// CVECatalog stores known vulnerability records indexed for rapid lookup.
type CVECatalog struct {
	mu      sync.RWMutex
	records map[string]*CVERecord
}

func NewCVECatalog() *CVECatalog {
	c := &CVECatalog{
		records: make(map[string]*CVERecord),
	}
	// Seed with notable high-impact reference vulnerabilities
	c.Add(&CVERecord{
		CVEID:        "CVE-2023-44487",
		Title:        "HTTP/2 Rapid Reset Attack",
		Description:  "The HTTP/2 protocol is susceptible to denial of service attacks via rapid stream reset frames.",
		CVSSScore:    7.5,
		EPSSScore:    0.97,
		InCISAKEV:    true,
		AffectedCPEs: []string{"cpe:2.3:a:golang:net", "cpe:2.3:a:nginx:nginx", "cpe:2.3:a:apache:http_server"},
	})
	c.Add(&CVERecord{
		CVEID:        "CVE-2023-4863",
		Title:        "libwebp Heap Buffer Overflow",
		Description:  "Heap buffer overflow in libwebp in Google Chrome, Chromium and webp libraries.",
		CVSSScore:    8.8,
		EPSSScore:    0.92,
		InCISAKEV:    true,
		AffectedCPEs: []string{"cpe:2.3:a:webmproject:libwebp"},
	})
	c.Add(&CVERecord{
		CVEID:        "CVE-2023-38606",
		Title:        "OpenSSH Remote Code Execution",
		Description:  "OpenSSH PKCS#11 provider memory corruption vulnerability.",
		CVSSScore:    9.8,
		EPSSScore:    0.85,
		InCISAKEV:    false,
		AffectedCPEs: []string{"cpe:2.3:a:openbsd:openssh:9.3"},
	})
	return c
}

func (c *CVECatalog) Add(record *CVERecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records[record.CVEID] = record
}

func (c *CVECatalog) MatchService(serviceName, version string) []*CVERecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	query := strings.ToLower(serviceName)
	var matches []*CVERecord
	for _, rec := range c.records {
		for _, cpe := range rec.AffectedCPEs {
			if strings.Contains(strings.ToLower(cpe), query) {
				matches = append(matches, rec)
				break
			}
		}
	}
	return matches
}
