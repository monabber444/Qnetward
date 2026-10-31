package securityengine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrVerificationNotFound = errors.New("asset verification record not found")
	ErrTokenMismatch        = errors.New("verification token mismatch")
	ErrVerificationPending  = errors.New("asset ownership verification is pending")
)

type VerificationMethod string

const (
	MethodDNSTXT        VerificationMethod = "dns_txt"
	MethodHTTPWellKnown VerificationMethod = "http_wellknown"
)

type VerificationStatus string

const (
	StatusPending  VerificationStatus = "pending"
	StatusVerified VerificationStatus = "verified"
	StatusRevoked  VerificationStatus = "revoked"
)

type AssetVerification struct {
	ID         string             `json:"id"`
	TenantID   int64              `json:"tenant_id"`
	Target     string             `json:"target"` // Domain or IP
	Token      string             `json:"token"`
	Method     VerificationMethod `json:"method"`
	Status     VerificationStatus `json:"status"`
	CreatedAt  time.Time          `json:"created_at"`
	VerifiedAt *time.Time         `json:"verified_at,omitempty"`
}

type OwnershipVerifier struct {
	mu            sync.RWMutex
	records       map[string]*AssetVerification // "tenant:target" -> record
	lookupTXTFunc func(ctx context.Context, name string) ([]string, error)
	httpGetFunc   func(ctx context.Context, url string) (string, error)
}

func NewOwnershipVerifier() *OwnershipVerifier {
	return &OwnershipVerifier{
		records: make(map[string]*AssetVerification),
		lookupTXTFunc: func(ctx context.Context, name string) ([]string, error) {
			var r net.Resolver
			return r.LookupTXT(ctx, name)
		},
		httpGetFunc: func(ctx context.Context, url string) (string, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return "", err
			}
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	}
}

// SetLookupHooks allows overriding DNS and HTTP lookups for unit testing and offline CI.
func (v *OwnershipVerifier) SetLookupHooks(
	dnsHook func(ctx context.Context, name string) ([]string, error),
	httpHook func(ctx context.Context, url string) (string, error),
) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if dnsHook != nil {
		v.lookupTXTFunc = dnsHook
	}
	if httpHook != nil {
		v.httpGetFunc = httpHook
	}
}

func (v *OwnershipVerifier) GenerateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (v *OwnershipVerifier) RegisterPending(tenantID int64, target string, method VerificationMethod) *AssetVerification {
	v.mu.Lock()
	defer v.mu.Unlock()

	target = strings.ToLower(strings.TrimSpace(target))
	key := fmt.Sprintf("%d:%s", tenantID, target)

	token := v.GenerateToken()
	rec := &AssetVerification{
		ID:        fmt.Sprintf("ver-%d", time.Now().UnixNano()),
		TenantID:  tenantID,
		Target:    target,
		Token:     token,
		Method:    method,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}

	v.records[key] = rec
	return rec
}

func (v *OwnershipVerifier) GetVerification(tenantID int64, target string) (*AssetVerification, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	target = strings.ToLower(strings.TrimSpace(target))
	key := fmt.Sprintf("%d:%s", tenantID, target)
	rec, ok := v.records[key]
	return rec, ok
}

// IsVerified checks if target has an active verified ownership status.
func (v *OwnershipVerifier) IsVerified(tenantID int64, target string) bool {
	rec, ok := v.GetVerification(tenantID, target)
	return ok && rec.Status == StatusVerified
}

// Verify executes the automated verification check against the target.
func (v *OwnershipVerifier) Verify(ctx context.Context, tenantID int64, target string) error {
	v.mu.RLock()
	target = strings.ToLower(strings.TrimSpace(target))
	key := fmt.Sprintf("%d:%s", tenantID, target)
	rec, ok := v.records[key]
	lookupTXT := v.lookupTXTFunc
	httpGet := v.httpGetFunc
	v.mu.RUnlock()

	if !ok {
		return ErrVerificationNotFound
	}

	expectedPrefix := "qnetward-verify=" + rec.Token

	switch rec.Method {
	case MethodDNSTXT:
		txts, err := lookupTXT(ctx, rec.Target)
		if err != nil {
			return fmt.Errorf("DNS TXT query failed: %w", err)
		}
		found := false
		for _, txt := range txts {
			if strings.TrimSpace(txt) == expectedPrefix || strings.Contains(txt, rec.Token) {
				found = true
				break
			}
		}
		if !found {
			return ErrTokenMismatch
		}

	case MethodHTTPWellKnown:
		url := fmt.Sprintf("http://%s/.well-known/qnetward-verify.txt", rec.Target)
		content, err := httpGet(ctx, url)
		if err != nil {
			return fmt.Errorf("HTTP token fetch failed: %w", err)
		}
		if !strings.Contains(content, rec.Token) {
			return ErrTokenMismatch
		}

	default:
		return fmt.Errorf("unsupported verification method %s", rec.Method)
	}

	v.mu.Lock()
	now := time.Now().UTC()
	rec.Status = StatusVerified
	rec.VerifiedAt = &now
	v.mu.Unlock()
	return nil
}
