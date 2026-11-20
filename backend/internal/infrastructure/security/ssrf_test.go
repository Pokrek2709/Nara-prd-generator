package security_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/your-org/prd-generator/backend/internal/infrastructure/security"
)

func TestValidateSafeBaseURL(t *testing.T) {
	mockResolver := func(ctx context.Context, network, host string) ([]net.IP, error) {
		switch host {
		case "api.openai.com":
			return []net.IP{net.ParseIP("104.18.7.192")}, nil
		case "api.deepseek.com":
			return []net.IP{net.ParseIP("104.18.28.140")}, nil
		case "internal-spoof.com":
			return []net.IP{net.ParseIP("192.168.1.100")}, nil
		default:
			return nil, errors.New("host not found")
		}
	}

	tests := []struct {
		name           string
		url            string
		allowLocalhost bool
		expectErr      error
	}{
		{
			name:           "Empty URL",
			url:            "",
			allowLocalhost: false,
			expectErr:      nil,
		},
		{
			name:           "Valid public HTTPS URL with mock resolver",
			url:            "https://api.openai.com/v1",
			allowLocalhost: false,
			expectErr:      nil,
		},
		{
			name:           "Valid public HTTP URL with direct public IP",
			url:            "http://93.184.216.34/v1",
			allowLocalhost: false,
			expectErr:      nil,
		},
		{
			name:           "DNS spoof resolving to private IP",
			url:            "https://internal-spoof.com/api",
			allowLocalhost: false,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Unsupported file scheme",
			url:            "file:///etc/passwd",
			allowLocalhost: false,
			expectErr:      security.ErrUnsupportedURL,
		},
		{
			name:           "Unsupported gopher scheme",
			url:            "gopher://127.0.0.1:70",
			allowLocalhost: false,
			expectErr:      security.ErrUnsupportedURL,
		},
		{
			name:           "Disallowed loopback in prod",
			url:            "http://127.0.0.1:8000/v1",
			allowLocalhost: false,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Disallowed localhost host in prod",
			url:            "http://localhost:11434/v1",
			allowLocalhost: false,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Allowed localhost when allowLocalhost is true",
			url:            "http://localhost:11434/v1",
			allowLocalhost: true,
			expectErr:      nil,
		},
		{
			name:           "Allowed 127.0.0.1 when allowLocalhost is true",
			url:            "http://127.0.0.1:11434/v1",
			allowLocalhost: true,
			expectErr:      nil,
		},
		{
			name:           "Cloud metadata IP 169.254.169.254 blocked even if allowLocalhost true",
			url:            "http://169.254.169.254/latest/meta-data/",
			allowLocalhost: true,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Cloud metadata hostname blocked",
			url:            "http://metadata.google.internal/computeMetadata/v1/",
			allowLocalhost: true,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Private 10.x.x.x IP blocked",
			url:            "http://10.0.0.5:8000/v1",
			allowLocalhost: true,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Private 192.168.x.x IP blocked",
			url:            "http://192.168.1.10:8000/v1",
			allowLocalhost: true,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Private 172.16.x.x IP blocked",
			url:            "http://172.16.0.2:8000/v1",
			allowLocalhost: true,
			expectErr:      security.ErrSSRFBlocked,
		},
		{
			name:           "Unspecified 0.0.0.0 blocked",
			url:            "http://0.0.0.0:8000/v1",
			allowLocalhost: true,
			expectErr:      security.ErrSSRFBlocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidateSafeBaseURLWithResolver(tt.url, tt.allowLocalhost, mockResolver)
			if tt.expectErr == nil {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectErr)
				}
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("Expected error %v, got: %v", tt.expectErr, err)
				}
			}
		})
	}
}

func TestSafeDialContextRejectsPrivateResolvedIPBeforeDial(t *testing.T) {
	dialer := &net.Dialer{}
	dial := security.SafeDialContextWithResolver(false, func(context.Context, string, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("192.168.1.20")}, nil
	}, dialer)
	_, err := dial(context.Background(), "tcp", "provider.example:443")
	if !errors.Is(err, security.ErrSSRFBlocked) {
		t.Fatalf("expected private resolved IP to be blocked, got %v", err)
	}
}
