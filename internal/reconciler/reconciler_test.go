package reconciler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseRetryAfter_IntegerSeconds(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{"valid seconds", "30", 30 * time.Second},
		{"one second", "1", 1 * time.Second},
		{"large value", "120", 120 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.value)
			if got != tt.expected {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestParseRetryAfter_EmptyAndInvalid(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{"empty string", "", 60 * time.Second},
		{"invalid string", "not-a-number", 60 * time.Second},
		{"zero seconds", "0", 60 * time.Second},
		{"negative seconds", "-5", 60 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.value)
			if got != tt.expected {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestRateLimitingRoundTripper_PassesThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := &RateLimitingRoundTripper{Delegate: http.DefaultTransport}
	req, _ := http.NewRequest("GET", server.URL, nil)

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRateLimitingRoundTripper_Returns429Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "45")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	rt := &RateLimitingRoundTripper{Delegate: http.DefaultTransport}
	req, _ := http.NewRequest("GET", server.URL, nil)

	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}

	rateLimitErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rateLimitErr.RetryAfter != 45*time.Second {
		t.Errorf("expected RetryAfter=45s, got %v", rateLimitErr.RetryAfter)
	}
}

func TestNewRateLimitingTransport_NilDefault(t *testing.T) {
	rt := NewRateLimitingTransport(nil)
	rlrt, ok := rt.(*RateLimitingRoundTripper)
	if !ok {
		t.Fatal("expected *RateLimitingRoundTripper")
	}
	if rlrt.Delegate != http.DefaultTransport {
		t.Error("expected default transport when nil is passed")
	}
}

func TestBackoffConfig_NextBackoff(t *testing.T) {
	cfg := DefaultBackoffConfig()

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
	}

	for _, tt := range tests {
		got := cfg.NextBackoff(tt.attempt)
		if got != tt.expected {
			t.Errorf("NextBackoff(%d) = %v, want %v", tt.attempt, got, tt.expected)
		}
	}
}

func TestBackoffConfig_NextBackoff_CapsAtMax(t *testing.T) {
	cfg := BackoffConfig{
		Base:   1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2.0,
	}

	// attempt 4 would be 16s without cap, should be capped at 10s
	got := cfg.NextBackoff(4)
	if got != 10*time.Second {
		t.Errorf("NextBackoff(4) = %v, want %v (capped)", got, 10*time.Second)
	}
}

func TestRateLimitError_Error(t *testing.T) {
	err := &RateLimitError{RetryAfter: 30 * time.Second}
	expected := "linear API rate limited: retry after 30s"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
