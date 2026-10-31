package ctwatcher

import (
	"context"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"kitten", "sitting", 3},
		{"qnetward", "qnetward", 0},
		{"qnetward", "qnetward1", 1},
		{"qnetward", "gnetward", 1},
		{"", "test", 4},
		{"test", "", 4},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestBrandDetector_PhishingVariants(t *testing.T) {
	detector := NewBrandDetector(WatcherConfig{
		Brands:              []string{"qnetward.com"},
		SimilarityThreshold: 0.75,
		CheckHomoglyphs:     true,
		CheckSubdomains:     true,
	})

	cert := CTCertificate{
		Issuer:    "Let's Encrypt Authority X3",
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(90 * 24 * time.Hour),
	}

	// 1. Identical brand domain should not be flagged
	if f := detector.EvaluateDomain("qnetward.com", cert); f != nil {
		t.Errorf("Expected nil finding for identical brand domain, got %+v", f)
	}

	// 2. Homoglyph confusable: "vv" instead of "w"
	fHomoglyph := detector.EvaluateDomain("qnetvvard.com", cert)
	if fHomoglyph == nil || fHomoglyph.DetectionType != DetectHomoglyph {
		t.Fatalf("Expected DetectHomoglyph for qnetvvard.com, got %+v", fHomoglyph)
	}
	if fHomoglyph.Severity != "HIGH" {
		t.Errorf("Expected HIGH severity, got %s", fHomoglyph.Severity)
	}

	// 3. Subdomain impersonation
	fSub := detector.EvaluateDomain("qnetward.secure-portal.net", cert)
	if fSub == nil || fSub.DetectionType != DetectSubdomainImpersonation {
		t.Fatalf("Expected DetectSubdomainImpersonation for qnetward.secure-portal.net, got %+v", fSub)
	}

	// 4. Typo / 1-edit distance
	fTypo := detector.EvaluateDomain("qnetwardx.com", cert)
	if fTypo == nil || fTypo.DetectionType != DetectLevenshtein {
		t.Fatalf("Expected DetectLevenshtein for qnetwardx.com, got %+v", fTypo)
	}
	if fTypo.SimilarityScore < 0.85 {
		t.Errorf("Expected high similarity score for 1 edit, got %f", fTypo.SimilarityScore)
	}
}

func TestStreamConsumer_EventBusPublishing(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	detector := NewBrandDetector(WatcherConfig{
		Brands:              []string{"qnetward.com"},
		SimilarityThreshold: 0.75,
		CheckHomoglyphs:     true,
		CheckSubdomains:     true,
	})

	consumer := NewStreamConsumer(detector, bus)

	var wg sync.WaitGroup
	var receivedSub int
	var mu sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wg.Add(1)
	_, err := bus.Subscribe(ctx, TopicTyposquatFinding, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		receivedSub++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	cert := CTCertificate{
		Domain: "legit-partner.org",
		SANs:   []string{"qnetvvard.com", "other-site.com"},
		Issuer: "ZeroSSL RSA Domain Secure Site CA",
	}

	findings, err := consumer.ProcessCertificate(ctx, cert)
	if err != nil {
		t.Fatalf("ProcessCertificate failed: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("Expected 1 finding, got %d", len(findings))
	}
	if findings[0].SuspectDomain != "qnetvvard.com" {
		t.Errorf("Expected finding for qnetvvard.com, got %s", findings[0].SuspectDomain)
	}

	wg.Wait()

	mu.Lock()
	if receivedSub != 1 {
		t.Errorf("Expected 1 event bus message, got %d", receivedSub)
	}
	mu.Unlock()

	stored := consumer.GetFindings()
	if len(stored) != 1 {
		t.Errorf("Expected 1 stored finding, got %d", len(stored))
	}

	consumer.ClearFindings()
	if len(consumer.GetFindings()) != 0 {
		t.Errorf("Expected 0 findings after clear, got %d", len(consumer.GetFindings()))
	}
}
