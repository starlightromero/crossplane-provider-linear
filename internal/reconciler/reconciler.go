// Package reconciler documents and configures the reconciliation behavior for
// provider-linear managed resources.
//
// # Reconciliation Overview (Requirements 12.1–12.7)
//
// All managed resources follow the standard Crossplane/Upjet reconciliation loop:
//
//  1. Observe: Read current state from Linear API via Terraform Read
//  2. Compare: Diff observed state against desired spec
//  3. Act: Create, Update, or Delete via Terraform operations
//  4. Report: Set Ready and Synced conditions; populate status.atProvider
//
// # Status Conditions (Requirements 12.1, 12.2)
//
// The Upjet framework automatically manages status conditions:
//   - Ready=True is set after a successful create or update operation completes
//     and the resource is confirmed to exist in the Linear API.
//   - Synced=True is set when the observed state from the Linear API matches
//     the desired state in spec.forProvider.
//   - Synced=False is set when there is a drift between desired and observed
//     state, or when an error occurs during reconciliation.
//
// No custom code is needed — Upjet's managed reconciler handles this via the
// crossplane-runtime's managed.Reconciler condition-setting logic.
//
// # status.atProvider Population (Requirement 12.7)
//
// Upjet automatically populates status.atProvider by reading the Terraform
// state after each successful observe/create/update operation. The Terraform
// provider's Read function returns the full resource state from the Linear
// GraphQL API, and Upjet maps those fields into the CRD's status.atProvider
// section. This includes computed fields like Linear-assigned UUIDs (e.g.,
// Team.status.atProvider.id).
//
// # Exponential Backoff (Requirement 12.3)
//
// When the Linear API is unreachable or returns server errors (5xx), the
// Crossplane runtime's managed reconciler applies exponential backoff
// automatically. The backoff is configured via the controller-runtime's
// rate limiter. Additionally, this package provides a custom rate limiter
// that respects Linear API 429 responses.
//
// # Rate Limiting — 429 Retry-After (Requirement 12.4)
//
// The Linear API returns HTTP 429 with a Retry-After header when rate limits
// are exceeded. The RateLimitingRoundTripper in this package intercepts 429
// responses and returns a structured error that the Upjet reconciler uses to
// delay the next reconciliation attempt by the indicated duration.
//
// # Deletion Policy — Orphan (Requirement 12.5)
//
// The Crossplane runtime natively supports the deletionPolicy field on all
// managed resources. When set to "Orphan", the Kubernetes resource is removed
// but the corresponding Linear API object is left intact (no Terraform Destroy
// is called). This is handled entirely by the crossplane-runtime's managed
// reconciler — no custom code is needed.
//
// # Management Policies (Requirement 12.6)
//
// Management policies provide fine-grained control over which operations the
// provider performs (ObserveOnly, Create, Update, Delete, or combinations).
// This is enabled via the EnableAlphaManagementPolicies feature flag in
// cmd/provider/main.go (already set to default "true"). The Upjet framework
// respects management policies when they are enabled in the feature flags
// passed to the controller options.
package reconciler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// DefaultBackoffBase is the base duration for exponential backoff when the
// Linear API is unreachable or returns server errors.
const DefaultBackoffBase = 1 * time.Second

// DefaultBackoffMax is the maximum backoff duration before retrying.
const DefaultBackoffMax = 5 * time.Minute

// DefaultBackoffFactor is the multiplier applied to the backoff duration
// after each consecutive failure.
const DefaultBackoffFactor = 2.0

// RateLimitError is returned when the Linear API responds with HTTP 429.
// It carries the retry delay parsed from the Retry-After header.
type RateLimitError struct {
	// RetryAfter is the duration the client should wait before retrying.
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("linear API rate limited: retry after %s", e.RetryAfter)
}

// RateLimitingRoundTripper wraps an http.RoundTripper and intercepts HTTP 429
// responses from the Linear API. When a 429 is received, it parses the
// Retry-After header and returns a RateLimitError so the caller (Upjet
// reconciler) can schedule the next reconciliation after the indicated delay.
type RateLimitingRoundTripper struct {
	// Delegate is the underlying HTTP transport.
	Delegate http.RoundTripper
}

// RoundTrip executes the HTTP request and checks for rate limit responses.
func (rt *RateLimitingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.Delegate.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return resp, &RateLimitError{RetryAfter: retryAfter}
	}

	return resp, nil
}

// parseRetryAfter parses the Retry-After header value. It supports both
// integer seconds and HTTP-date formats. If parsing fails, it returns a
// default delay of 60 seconds.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 60 * time.Second
	}

	// Try parsing as integer seconds first (most common for APIs).
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 60 * time.Second
		}
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP-date (RFC 1123).
	if t, err := http.ParseTime(value); err == nil {
		delay := time.Until(t)
		if delay <= 0 {
			return 60 * time.Second
		}
		return delay
	}

	// Fallback to default.
	return 60 * time.Second
}

// NewRateLimitingTransport wraps the given transport with rate limit detection
// for the Linear API. Use this when constructing the HTTP client passed to the
// Terraform provider configuration.
func NewRateLimitingTransport(transport http.RoundTripper) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &RateLimitingRoundTripper{Delegate: transport}
}

// BackoffConfig holds the exponential backoff parameters for retrying
// failed reconciliation attempts when the Linear API is unreachable.
type BackoffConfig struct {
	// Base is the initial backoff duration after the first failure.
	Base time.Duration

	// Max is the maximum backoff duration.
	Max time.Duration

	// Factor is the multiplier applied after each consecutive failure.
	Factor float64
}

// DefaultBackoffConfig returns the default exponential backoff configuration
// for Linear API error retries.
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		Base:   DefaultBackoffBase,
		Max:    DefaultBackoffMax,
		Factor: DefaultBackoffFactor,
	}
}

// NextBackoff calculates the next backoff duration given the current attempt
// number (0-indexed). The result is capped at BackoffConfig.Max.
func (c BackoffConfig) NextBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.Base
	}

	backoff := c.Base
	for i := 0; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * c.Factor)
		if backoff > c.Max {
			return c.Max
		}
	}
	return backoff
}
