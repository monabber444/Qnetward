package dast

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"netwatcher/pkg/eventbus"
)

func TestAuditor_HeadersAndCORS(t *testing.T) {
	// Setup a mock insecure server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Insecure response: No CSP, no XFO, no XCTO, reflects CORS with credentials, server version leaked
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Server", "Apache/2.4.52 (Ubuntu)")
		w.Header().Set("X-Powered-By", "Express")

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "secret123",
			Secure:   false,
			HttpOnly: false,
		})

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello Insecure World"))
	}))
	defer ts.Close()

	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	auditor := NewAuditor(ts.Client(), bus)

	var busFindingsCount int
	var mu sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := bus.Subscribe(ctx, TopicDASTFinding, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		busFindingsCount++
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	cfg := AuditConfig{
		CheckHeaders:     true,
		CheckCORS:        true,
		CheckCookies:     true,
		CheckInfoLeakage: true,
	}

	report, err := auditor.AuditTarget(ctx, ts.URL, cfg)
	if err != nil {
		t.Fatalf("AuditTarget failed: %v", err)
	}

	if len(report.Findings) == 0 {
		t.Fatalf("Expected multiple findings on insecure server, got 0")
	}

	// Verify categories are present
	categories := make(map[string]bool)
	for _, f := range report.Findings {
		categories[f.Category] = true
	}

	if !categories["HEADER_MISSING"] {
		t.Errorf("Expected HEADER_MISSING finding")
	}
	if !categories["CORS_PERMISSIVE"] {
		t.Errorf("Expected CORS_PERMISSIVE finding")
	}
	if !categories["INFO_LEAKAGE"] {
		t.Errorf("Expected INFO_LEAKAGE finding")
	}

	reports := auditor.GetReports()
	if len(reports) != 1 {
		t.Errorf("Expected 1 report in auditor cache, got %d", len(reports))
	}
}

func TestOASTManager_TokenAndInteraction(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()

	mgr := NewOASTManager("oast.test.internal", bus)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var receivedCount int
	var mu sync.Mutex

	wg.Add(1)
	_, err := bus.Subscribe(ctx, TopicOASTInteraction, func(ctx context.Context, msg eventbus.Message) error {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to OAST topic: %v", err)
	}

	token, callbackDomain := mgr.GenerateToken("https://target.corp/api/search", "BLIND_SSRF")
	if token == "" || callbackDomain == "" {
		t.Fatalf("Failed to generate token or domain")
	}

	// Simulate inbound interaction via HTTPHandler
	handler := mgr.HTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "http://oast.test.internal/?token="+token, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK from OAST handler, got %d", rr.Code)
	}

	wg.Wait()

	mu.Lock()
	if receivedCount != 1 {
		t.Errorf("Expected 1 interaction published to bus, got %d", receivedCount)
	}
	mu.Unlock()

	interactions := mgr.GetInteractions()
	if len(interactions) != 1 {
		t.Fatalf("Expected 1 interaction recorded, got %d", len(interactions))
	}
	if interactions[0].Token != token {
		t.Errorf("Expected token %s, got %s", token, interactions[0].Token)
	}
	if interactions[0].PayloadType != "BLIND_SSRF" {
		t.Errorf("Expected BLIND_SSRF payload type, got %s", interactions[0].PayloadType)
	}
}
