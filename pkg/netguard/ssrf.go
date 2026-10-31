package netguard

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	ErrEmptyTarget       = errors.New("target cannot be empty")
	ErrTargetTooLong     = errors.New("target exceeds maximum length")
	ErrBlockedIP         = errors.New("target resolves to a blocked or restricted IP address")
	ErrLoopbackBlocked   = errors.New("loopback addresses are blocked")
	ErrMetadataBlocked   = errors.New("cloud metadata services are blocked")
	ErrLinkLocalBlocked  = errors.New("link-local addresses are blocked")
	ErrPrivateBlocked    = errors.New("private network addresses are blocked")
	ErrMulticastBlocked  = errors.New("multicast addresses are blocked")
	ErrUnspecifiedBlocked = errors.New("unspecified addresses are blocked")
)

type Options struct {
	// AllowPrivate allows RFC1918 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
	// and IPv6 Unique Local Addresses (fc00::/7) for internal enterprise networks.
	AllowPrivate bool

	// AllowLoopback allows 127.0.0.0/8 and ::1 (for local testing).
	AllowLoopback bool
}

// DefaultOptions provides strict anti-SSRF protection suitable for SaaS.
var DefaultOptions = Options{
	AllowPrivate:  false,
	AllowLoopback: false,
}

var (
	cloudMetadataIPv4 = net.ParseIP("169.254.169.254")
	cloudMetadataIPv6 = net.ParseIP("fd00:ec2::254")

	_, linkLocalIPv4Net, _     = net.ParseCIDR("169.254.0.0/16")
	_, linkLocalIPv6Net, _     = net.ParseCIDR("fe80::/10")
	_, cgnatNet, _             = net.ParseCIDR("100.64.0.0/10")
	_, documentationIPv4_1, _  = net.ParseCIDR("192.0.2.0/24")
	_, documentationIPv4_2, _  = net.ParseCIDR("198.51.100.0/24")
	_, documentationIPv4_3, _  = net.ParseCIDR("203.0.113.0/24")
	_, documentationIPv6, _    = net.ParseCIDR("2001:db8::/32")
	_, benchmarkNet, _         = net.ParseCIDR("198.18.0.0/15")

	_, private10Net, _  = net.ParseCIDR("10.0.0.0/8")
	_, private172Net, _ = net.ParseCIDR("172.16.0.0/12")
	_, private192Net, _ = net.ParseCIDR("192.168.0.0/16")
	_, privateULANet, _ = net.ParseCIDR("fc00::/7")
)

// ValidateIP verifies that an individual IP does not violate SSRF rules.
func ValidateIP(ip net.IP, opts Options) error {
	if ip == nil {
		return ErrEmptyTarget
	}

	// Normalize IPv4-mapped IPv6 addresses (e.g. ::ffff:127.0.0.1)
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}

	// Unspecified / 0.0.0.0 / ::
	if ip.IsUnspecified() {
		return ErrUnspecifiedBlocked
	}

	// Loopback
	if ip.IsLoopback() {
		if !opts.AllowLoopback {
			return ErrLoopbackBlocked
		}
		return nil
	}

	// Multicast
	if ip.IsMulticast() {
		return ErrMulticastBlocked
	}

	// Cloud metadata (AWS, GCP, Azure, OpenStack)
	if ip.Equal(cloudMetadataIPv4) || ip.Equal(cloudMetadataIPv6) {
		return ErrMetadataBlocked
	}

	// Link-local
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || linkLocalIPv4Net.Contains(ip) || linkLocalIPv6Net.Contains(ip) {
		return ErrLinkLocalBlocked
	}

	// CGNAT (RFC 6598) - 100.64.0.0/10
	if cgnatNet.Contains(ip) {
		return ErrBlockedIP
	}

	// Documentation & benchmark ranges
	if documentationIPv4_1.Contains(ip) || documentationIPv4_2.Contains(ip) || documentationIPv4_3.Contains(ip) ||
		documentationIPv6.Contains(ip) || benchmarkNet.Contains(ip) {
		return ErrBlockedIP
	}

	// Private ranges (RFC 1918 & RFC 4193 ULA)
	if isPrivateNetwork(ip) {
		if !opts.AllowPrivate {
			return ErrPrivateBlocked
		}
	}

	return nil
}

// ValidateTarget resolves host/domain/IP and checks all resolved addresses against SSRF policies.
func ValidateTarget(ctx context.Context, target string, opts Options) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return ErrEmptyTarget
	}
	if len(target) > 255 {
		return ErrTargetTooLong
	}

	// If host contains port, strip it for validation
	if host, _, err := net.SplitHostPort(target); err == nil {
		target = host
	}
	target = strings.Trim(target, "[]")

	// Direct IP check
	if ip := net.ParseIP(target); ip != nil {
		return ValidateIP(ip, opts)
	}

	// Block localhost and standard loopback hostnames directly
	lower := strings.ToLower(target)
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") || strings.HasSuffix(lower, ".local") {
		if !opts.AllowLoopback {
			return ErrLoopbackBlocked
		}
	}

	// Resolve hostname
	resolver := net.DefaultResolver
	ips, err := resolver.LookupIPAddr(ctx, target)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses found for target %q", target)
	}

	// All resolved IPs must pass validation to prevent DNS rebinding
	for _, ipAddr := range ips {
		if err := ValidateIP(ipAddr.IP, opts); err != nil {
			return fmt.Errorf("resolved IP %s: %w", ipAddr.IP.String(), err)
		}
	}

	return nil
}

func isPrivateNetwork(ip net.IP) bool {
	if ip.IsPrivate() {
		return true
	}
	return private10Net.Contains(ip) ||
		private172Net.Contains(ip) ||
		private192Net.Contains(ip) ||
		privateULANet.Contains(ip)
}
