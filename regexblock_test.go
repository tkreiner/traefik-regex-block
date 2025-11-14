package traefik_regex_block

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNew validates the New function creates a valid plugin
func TestNew(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`, `/admin`}
	cfg.BlockDurationMinutes = 60
	cfg.EnableDebug = true

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	if handler == nil {
		t.Fatal("Handler should not be nil")
	}
}

// TestNewWithInvalidRegex validates that invalid regex patterns are handled gracefully
func TestNewWithInvalidRegex(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`[invalid(regex`}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err == nil {
		t.Fatal("Expected error for invalid regex patterns")
	}
	if handler != nil {
		t.Fatal("Handler should be nil when no valid patterns")
	}
}

// TestNewWithMixedValidInvalidRegex validates mixed valid/invalid patterns
func TestNewWithMixedValidInvalidRegex(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`[invalid(regex`, `/\.env`}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Should not error with at least one valid pattern: %v", err)
	}
	if handler == nil {
		t.Fatal("Handler should not be nil with valid patterns")
	}
}

// TestServeHTTP_MalformedRemoteAddr validates handling of malformed RemoteAddr
func TestServeHTTP_MalformedRemoteAddr(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Test with malformed IP (no port)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1"
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status OK for non-matching path, got %d", rw.Code)
	}
}

// TestServeHTTP_InvalidIP validates handling of invalid IP addresses
func TestServeHTTP_InvalidIP(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Test with invalid IP
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "invalid-ip-address"
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status OK when IP cannot be parsed, got %d", rw.Code)
	}
}

// TestServeHTTP_BlocksMatchingPattern validates blocking of matching patterns
func TestServeHTTP_BlocksMatchingPattern(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Test with matching path
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("Expected status NotFound for matching path, got %d", rw.Code)
	}
}

// TestServeHTTP_AllowsNonMatchingPattern validates allowing non-matching patterns
func TestServeHTTP_AllowsNonMatchingPattern(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Test with non-matching path
	req := httptest.NewRequest(http.MethodGet, "/normal-path", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status OK for non-matching path, got %d", rw.Code)
	}
}

// TestServeHTTP_WhitelistedIP validates whitelist functionality
func TestServeHTTP_WhitelistedIP(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}
	cfg.Whitelist = []string{"192.168.1.1"}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Test with whitelisted IP and matching path
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status OK for whitelisted IP, got %d", rw.Code)
	}
}

// TestServeHTTP_WhitelistedCIDR validates CIDR whitelist functionality
func TestServeHTTP_WhitelistedCIDR(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}
	cfg.Whitelist = []string{"192.168.0.0/16"}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Test with IP in whitelisted CIDR range
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req.RemoteAddr = "192.168.100.50:12345"
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status OK for IP in whitelisted CIDR, got %d", rw.Code)
	}
}

// TestServeHTTP_IPv6Support validates IPv6 address handling
func TestServeHTTP_IPv6Support(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}
	cfg.Whitelist = []string{"::1"}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Test with IPv6 loopback
	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req.RemoteAddr = "[::1]:12345"
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status OK for whitelisted IPv6, got %d", rw.Code)
	}
}

// TestServeHTTP_BlockedIPStaysBlocked validates that blocked IPs remain blocked
func TestServeHTTP_BlockedIPStaysBlocked(t *testing.T) {
	cfg := CreateConfig()
	cfg.RegexPatterns = []string{`/\.env`}
	cfg.BlockDurationMinutes = 60

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// First request - triggers block
	req1 := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	rw1 := httptest.NewRecorder()
	handler.ServeHTTP(rw1, req1)

	if rw1.Code != http.StatusNotFound {
		t.Errorf("Expected status NotFound for first blocked request, got %d", rw1.Code)
	}

	// Second request - should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/normal-path", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	rw2 := httptest.NewRecorder()
	handler.ServeHTTP(rw2, req2)

	if rw2.Code != http.StatusForbidden {
		t.Errorf("Expected status Forbidden for subsequent request from blocked IP, got %d", rw2.Code)
	}
}
