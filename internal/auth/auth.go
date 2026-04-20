// Package auth provides authentication credential resolution for the Linear API.
package auth

import (
	"context"
	"fmt"
	"time"
)

// CredentialSource defines the supported authentication source types.
type CredentialSource string

const (
	// SourceSecret uses a static API token from a Kubernetes Secret.
	SourceSecret CredentialSource = "Secret"
	// SourceOAuth2ClientCredentials uses OAuth2 client credentials flow.
	SourceOAuth2ClientCredentials CredentialSource = "OAuth2ClientCredentials"
	// SourceOAuth2 uses OAuth2 authorization code flow with refresh.
	SourceOAuth2 CredentialSource = "OAuth2"
)

// LinearTokenEndpoint is the OAuth2 token endpoint for Linear.
const LinearTokenEndpoint = "https://api.linear.app/oauth/token"

// DefaultScope is the default OAuth2 scope for client credentials.
const DefaultScope = "read,write"

// DefaultClientCredentialsLifetime is the token lifetime for client credentials (30 days).
const DefaultClientCredentialsLifetime = 30 * 24 * time.Hour

// DefaultAuthCodeLifetime is the token lifetime for authorization code tokens (24 hours).
const DefaultAuthCodeLifetime = 24 * time.Hour

// SecretRef identifies a Kubernetes Secret and key to read credentials from.
type SecretRef struct {
	Namespace string
	Name      string
	Key       string
}

// Credentials holds the resolved authentication configuration from a ProviderConfig.
type Credentials struct {
	Source    CredentialSource
	SecretRef SecretRef
	Scope     string // OAuth2ClientCredentials only; defaults to "read,write"
}

// SecretData holds key-value pairs read from a Kubernetes Secret.
type SecretData map[string][]byte

// SecretReader reads data from a Kubernetes Secret.
type SecretReader interface {
	ReadSecret(ctx context.Context, namespace, name string) (SecretData, error)
}

// SecretWriter writes data to a Kubernetes Secret.
type SecretWriter interface {
	WriteSecret(ctx context.Context, namespace, name string, data SecretData) error
}

// TokenProvider provides Bearer tokens for Linear API authentication.
type TokenProvider interface {
	// GetToken returns a valid Bearer token. It may perform token exchange
	// or refresh as needed.
	GetToken(ctx context.Context) (string, error)

	// InvalidateToken marks the current token as invalid (e.g., after a 401).
	// The next call to GetToken will obtain a fresh token.
	InvalidateToken()
}

// TokenResponse represents the response from the Linear OAuth2 token endpoint.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// HTTPTokenExchanger performs HTTP token exchange requests against the OAuth2 endpoint.
type HTTPTokenExchanger interface {
	ExchangeToken(ctx context.Context, params map[string]string) (*TokenResponse, error)
}

// ConditionSetter sets status conditions on a ProviderConfig.
type ConditionSetter interface {
	SetNotReady(reason, message string)
}

// AuthError represents an authentication error with a reason suitable for status conditions.
type AuthError struct {
	Reason  string
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Reason, e.Message)
}

// NewAuthError creates a new AuthError.
func NewAuthError(reason, message string) *AuthError {
	return &AuthError{Reason: reason, Message: message}
}
