package netguard

import (
	"context"
	"net"
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		opts      Options
		wantError error
	}{
		{
			name:      "public ipv4 google dns",
			ip:        "8.8.8.8",
			opts:      DefaultOptions,
			wantError: nil,
		},
		{
			name:      "public ipv4 cloudflare",
			ip:        "1.1.1.1",
			opts:      DefaultOptions,
			wantError: nil,
		},
		{
			name:      "public ipv6 google dns",
			ip:        "2001:4860:4860::8888",
			opts:      DefaultOptions,
			wantError: nil,
		},
		{
			name:      "loopback ipv4 blocked by default",
			ip:        "127.0.0.1",
			opts:      DefaultOptions,
			wantError: ErrLoopbackBlocked,
		},
		{
			name:      "loopback ipv4 custom 127.100.200.1",
			ip:        "127.100.200.1",
			opts:      DefaultOptions,
			wantError: ErrLoopbackBlocked,
		},
		{
			name:      "loopback ipv4 allowed when configured",
			ip:        "127.0.0.1",
			opts:      Options{AllowLoopback: true},
			wantError: nil,
		},
		{
			name:      "loopback ipv6 blocked by default",
			ip:        "::1",
			opts:      DefaultOptions,
			wantError: ErrLoopbackBlocked,
		},
		{
			name:      "cloud metadata ipv4 blocked",
			ip:        "169.254.169.254",
			opts:      DefaultOptions,
			wantError: ErrMetadataBlocked,
		},
		{
			name:      "cloud metadata ipv6 blocked",
			ip:        "fd00:ec2::254",
			opts:      DefaultOptions,
			wantError: ErrMetadataBlocked,
		},
		{
			name:      "link-local ipv4 blocked",
			ip:        "169.254.1.1",
			opts:      DefaultOptions,
			wantError: ErrLinkLocalBlocked,
		},
		{
			name:      "link-local ipv6 blocked",
			ip:        "fe80::1",
			opts:      DefaultOptions,
			wantError: ErrLinkLocalBlocked,
		},
		{
			name:      "private 10.x blocked by default",
			ip:        "10.0.1.5",
			opts:      DefaultOptions,
			wantError: ErrPrivateBlocked,
		},
		{
			name:      "private 172.16.x blocked by default",
			ip:        "172.16.0.1",
			opts:      DefaultOptions,
			wantError: ErrPrivateBlocked,
		},
		{
			name:      "private 192.168.x blocked by default",
			ip:        "192.168.1.1",
			opts:      DefaultOptions,
			wantError: ErrPrivateBlocked,
		},
		{
			name:      "private 192.168.x allowed when enabled",
			ip:        "192.168.1.1",
			opts:      Options{AllowPrivate: true},
			wantError: nil,
		},
		{
			name:      "unspecified 0.0.0.0 blocked",
			ip:        "0.0.0.0",
			opts:      DefaultOptions,
			wantError: ErrUnspecifiedBlocked,
		},
		{
			name:      "cgnat 100.64.0.1 blocked",
			ip:        "100.64.0.1",
			opts:      DefaultOptions,
			wantError: ErrBlockedIP,
		},
		{
			name:      "multicast ipv4 blocked",
			ip:        "224.0.0.1",
			opts:      DefaultOptions,
			wantError: ErrMulticastBlocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := net.ParseIP(tt.ip)
			if parsed == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			err := ValidateIP(parsed, tt.opts)
			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			}
		})
	}
}

func TestValidateTarget(t *testing.T) {
	ctx := context.Background()

	t.Run("empty target", func(t *testing.T) {
		err := ValidateTarget(ctx, "", DefaultOptions)
		if err != ErrEmptyTarget {
			t.Fatalf("expected ErrEmptyTarget, got %v", err)
		}
	})

	t.Run("localhost blocked", func(t *testing.T) {
		err := ValidateTarget(ctx, "localhost", DefaultOptions)
		if err == nil {
			t.Fatal("expected error for localhost, got nil")
		}
	})

	t.Run("localhost allowed when configured", func(t *testing.T) {
		err := ValidateTarget(ctx, "127.0.0.1:8080", Options{AllowLoopback: true})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("metadata with port blocked", func(t *testing.T) {
		err := ValidateTarget(ctx, "169.254.169.254:80", DefaultOptions)
		if err != ErrMetadataBlocked {
			t.Fatalf("expected ErrMetadataBlocked, got %v", err)
		}
	})

	t.Run("public IP with port allowed", func(t *testing.T) {
		err := ValidateTarget(ctx, "8.8.8.8:53", DefaultOptions)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}
