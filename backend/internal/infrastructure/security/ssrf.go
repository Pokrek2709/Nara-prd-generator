package security

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidURL     = errors.New("invalid custom base URL format")
	ErrUnsupportedURL = errors.New("custom base URL scheme must be http or https")
	ErrSSRFBlocked    = errors.New("custom base URL target is blocked (private, loopback, or metadata address)")
	ErrResolutionFail = errors.New("could not resolve custom base URL host")
)

type IPResolver func(ctx context.Context, network, host string) ([]net.IP, error)

var DefaultResolver IPResolver = net.DefaultResolver.LookupIP

// SafeDialContext resolves the target once, rejects non-public addresses, and
// dials the checked IP directly so redirects and DNS changes cannot bypass the
// URL validation performed when provider settings are saved.
func SafeDialContext(allowLocalhost bool) func(context.Context, string, string) (net.Conn, error) {
	return SafeDialContextWithResolver(allowLocalhost, DefaultResolver, &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second})
}

// SafeDialContextWithResolver permits deterministic resolver injection in tests.
func SafeDialContextWithResolver(allowLocalhost bool, resolver IPResolver, dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	if resolver == nil {
		resolver = DefaultResolver
	}
	if dialer == nil {
		dialer = &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("invalid provider address: %w", err)
		}
		if host == "" {
			return nil, ErrInvalidURL
		}

		var ips []net.IP
		if parsedIP := net.ParseIP(host); parsedIP != nil {
			ips = []net.IP{parsedIP}
		} else {
			ips, err = resolver(ctx, "ip", host)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrResolutionFail, err)
			}
		}
		if len(ips) == 0 {
			return nil, ErrResolutionFail
		}
		for _, ip := range ips {
			if err := checkIPSafety(ip, allowLocalhost); err != nil {
				return nil, err
			}
		}

		var dialErr error
		for _, ip := range ips {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			dialErr = err
		}
		return nil, dialErr
	}
}

// ValidateSafeBaseURL verifies that a given rawURL is safe from SSRF attacks.
// If allowLocalhost is false, loopback addresses (127.0.0.0/8, ::1, localhost) are blocked.
// Link-local addresses (169.254.0.0/16, cloud metadata) and private network subnets are blocked.
func ValidateSafeBaseURL(rawURL string, allowLocalhost bool) error {
	return ValidateSafeBaseURLWithResolver(rawURL, allowLocalhost, DefaultResolver)
}

// ValidateSafeBaseURLWithResolver allows passing a custom IPResolver (useful for unit testing without live network).
func ValidateSafeBaseURLWithResolver(rawURL string, allowLocalhost bool, resolver IPResolver) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrUnsupportedURL
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return ErrInvalidURL
	}

	// Immediate block for known cloud metadata hostnames
	lowerHost := strings.ToLower(hostname)
	if lowerHost == "metadata.google.internal" || lowerHost == "metadata" || lowerHost == "instance-data" {
		return ErrSSRFBlocked
	}

	// Check if hostname is an explicit loopback name
	isLoopbackHost := lowerHost == "localhost" || lowerHost == "localhost.localdomain" || strings.HasSuffix(lowerHost, ".localhost")
	if isLoopbackHost {
		if !allowLocalhost {
			return ErrSSRFBlocked
		}
		return nil
	}

	// Resolve host IPs
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ips []net.IP
	if parsedIP := net.ParseIP(hostname); parsedIP != nil {
		ips = []net.IP{parsedIP}
	} else {
		if resolver == nil {
			resolver = DefaultResolver
		}
		resolved, err := resolver(ctx, "ip", hostname)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrResolutionFail, err)
		}
		ips = resolved
	}

	if len(ips) == 0 {
		return ErrResolutionFail
	}

	for _, ip := range ips {
		if err := checkIPSafety(ip, allowLocalhost); err != nil {
			return err
		}
	}

	return nil
}

func checkIPSafety(ip net.IP, allowLocalhost bool) error {
	if ip == nil || ip.IsMulticast() {
		return ErrSSRFBlocked
	}
	// 1. Unconditionally block AWS/GCP/Azure link-local metadata (169.254.169.254 and 169.254.0.0/16)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() {
		return ErrSSRFBlocked
	}
	if ip.Equal(net.ParseIP("169.254.169.254")) {
		return ErrSSRFBlocked
	}

	// 2. Unspecified addresses (0.0.0.0, ::)
	if ip.IsUnspecified() {
		return ErrSSRFBlocked
	}

	// 3. Loopback (127.0.0.0/8, ::1)
	if ip.IsLoopback() {
		if !allowLocalhost {
			return ErrSSRFBlocked
		}
		return nil
	}

	// 4. Private IP subnets (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7)
	if ip.IsPrivate() {
		return ErrSSRFBlocked
	}

	return nil
}
