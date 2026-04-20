//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/avodah-inc/provider-linear/internal/auth"
)

// --- Mock Linear GraphQL API Server ---

// GraphQLRequest represents a parsed GraphQL request body.
type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data   any            `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message string `json:"message"`
}

// MockLinearServer provides a configurable mock Linear GraphQL API for integration tests.
type MockLinearServer struct {
	mu       sync.Mutex
	server   *httptest.Server
	handlers map[string]GraphQLHandler
	requests []GraphQLRequest

	// RateLimitAfter triggers a 429 response after N requests. 0 means no rate limiting.
	RateLimitAfter int
	// RetryAfterSeconds is the Retry-After header value for 429 responses.
	RetryAfterSeconds int
	requestCount      int
}

// GraphQLHandler processes a GraphQL request and returns a response.
type GraphQLHandler func(req GraphQLRequest) GraphQLResponse

// NewMockLinearServer creates and starts a new mock Linear GraphQL API server.
func NewMockLinearServer(t *testing.T) *MockLinearServer {
	t.Helper()
	m := &MockLinearServer{
		handlers:          make(map[string]GraphQLHandler),
		RetryAfterSeconds: 60,
	}

	m.server = httptest.NewServer(http.HandlerFunc(m.handleRequest))
	t.Cleanup(m.server.Close)
	return m
}

// URL returns the base URL of the mock server.
func (m *MockLinearServer) URL() string {
	return m.server.URL
}

// RegisterHandler registers a handler for a specific operation name or query substring.
func (m *MockLinearServer) RegisterHandler(operationKey string, handler GraphQLHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[operationKey] = handler
}

// Requests returns all received GraphQL requests.
func (m *MockLinearServer) Requests() []GraphQLRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]GraphQLRequest{}, m.requests...)
}

// RequestCount returns the total number of requests received.
func (m *MockLinearServer) RequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requestCount
}

func (m *MockLinearServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requestCount++
	count := m.requestCount
	m.mu.Unlock()

	// Check rate limiting
	if m.RateLimitAfter > 0 && count > m.RateLimitAfter {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", m.RetryAfterSeconds))
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	// Parse request (allow empty body for simple connectivity checks)
	var req GraphQLRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Treat as empty query if body can't be parsed
			req = GraphQLRequest{}
		}
	}

	m.mu.Lock()
	m.requests = append(m.requests, req)
	m.mu.Unlock()

	// Find matching handler
	m.mu.Lock()
	var handler GraphQLHandler
	for key, h := range m.handlers {
		if containsSubstring(req.Query, key) {
			handler = h
			break
		}
	}
	m.mu.Unlock()

	if handler == nil {
		// Default: return empty data
		resp := GraphQLResponse{Data: map[string]any{}}
		writeJSON(w, resp)
		return
	}

	resp := handler(req)
	writeJSON(w, resp)
}

// --- Mock OAuth2 Token Server ---

// MockOAuth2Server provides a configurable mock OAuth2 token endpoint.
type MockOAuth2Server struct {
	mu     sync.Mutex
	server *httptest.Server

	// AccessToken is the token returned on successful exchange.
	AccessToken string
	// RefreshToken is the refresh token returned on successful exchange.
	RefreshToken string
	// ExpiresIn is the token lifetime in seconds.
	ExpiresIn int64
	// FailExchange causes the server to return an error on token exchange.
	FailExchange bool
	// Requests tracks all received token requests.
	Requests []map[string]string
}

// NewMockOAuth2Server creates and starts a new mock OAuth2 token server.
func NewMockOAuth2Server(t *testing.T) *MockOAuth2Server {
	t.Helper()
	m := &MockOAuth2Server{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresIn:    86400,
	}

	m.server = httptest.NewServer(http.HandlerFunc(m.handleToken))
	t.Cleanup(m.server.Close)
	return m
}

// URL returns the base URL of the mock OAuth2 server.
func (m *MockOAuth2Server) URL() string {
	return m.server.URL
}

func (m *MockOAuth2Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	params := make(map[string]string)
	for key := range r.Form {
		params[key] = r.FormValue(key)
	}

	m.mu.Lock()
	m.Requests = append(m.Requests, params)
	m.mu.Unlock()

	if m.FailExchange {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "invalid_grant"})
		return
	}

	resp := auth.TokenResponse{
		AccessToken:  m.AccessToken,
		RefreshToken: m.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    m.ExpiresIn,
	}
	writeJSON(w, resp)
}

// --- Mock K8s Secret Store ---

// MockSecretStore provides an in-memory K8s Secret store for integration tests.
type MockSecretStore struct {
	mu      sync.Mutex
	secrets map[string]auth.SecretData
}

// NewMockSecretStore creates a new in-memory secret store.
func NewMockSecretStore() *MockSecretStore {
	return &MockSecretStore{
		secrets: make(map[string]auth.SecretData),
	}
}

// SetSecret stores secret data for a given namespace/name.
func (s *MockSecretStore) SetSecret(namespace, name string, data auth.SecretData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := namespace + "/" + name
	s.secrets[key] = data
}

// ReadSecret implements auth.SecretReader.
func (s *MockSecretStore) ReadSecret(_ context.Context, namespace, name string) (auth.SecretData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := namespace + "/" + name
	data, ok := s.secrets[key]
	if !ok {
		return nil, fmt.Errorf("secret %s not found", key)
	}
	return data, nil
}

// WriteSecret implements auth.SecretWriter.
func (s *MockSecretStore) WriteSecret(_ context.Context, namespace, name string, data auth.SecretData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := namespace + "/" + name
	s.secrets[key] = data
	return nil
}

// GetSecret retrieves the current secret data (for test assertions).
func (s *MockSecretStore) GetSecret(namespace, name string) (auth.SecretData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := namespace + "/" + name
	data, ok := s.secrets[key]
	return data, ok
}

// --- Mock HTTP Token Exchanger ---

// MockHTTPExchanger wraps a mock OAuth2 server as an HTTPTokenExchanger.
type MockHTTPExchanger struct {
	mu        sync.Mutex
	responses []*auth.TokenResponse
	errors    []error
	callCount int
}

// NewMockHTTPExchanger creates a new mock exchanger with preset responses.
func NewMockHTTPExchanger(responses []*auth.TokenResponse, errors []error) *MockHTTPExchanger {
	return &MockHTTPExchanger{
		responses: responses,
		errors:    errors,
	}
}

// ExchangeToken implements auth.HTTPTokenExchanger.
func (m *MockHTTPExchanger) ExchangeToken(_ context.Context, _ map[string]string) (*auth.TokenResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := m.callCount
	m.callCount++

	if idx < len(m.errors) && m.errors[idx] != nil {
		return nil, m.errors[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return &auth.TokenResponse{AccessToken: "default-token", ExpiresIn: 3600}, nil
}

// --- Utility Functions ---

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
