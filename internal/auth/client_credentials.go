package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ClientCredentialsTokenProvider implements TokenProvider using OAuth2 client credentials flow.
type ClientCredentialsTokenProvider struct {
	mu            sync.Mutex
	reader        SecretReader
	exchanger     HTTPTokenExchanger
	secretRef     SecretRef
	scope         string
	token         string
	expiresAt     time.Time
	tokenLifetime time.Duration
}

// NewClientCredentialsTokenProvider creates a TokenProvider that exchanges client credentials
// for an access token via the Linear OAuth2 token endpoint.
func NewClientCredentialsTokenProvider(reader SecretReader, exchanger HTTPTokenExchanger, ref SecretRef, scope string) *ClientCredentialsTokenProvider {
	if scope == "" {
		scope = DefaultScope
	}
	return &ClientCredentialsTokenProvider{
		reader:        reader,
		exchanger:     exchanger,
		secretRef:     ref,
		scope:         scope,
		tokenLifetime: DefaultClientCredentialsLifetime,
	}
}

// GetToken returns a cached access token or exchanges client credentials for a new one.
func (p *ClientCredentialsTokenProvider) GetToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token != "" && time.Now().Before(p.expiresAt) {
		return p.token, nil
	}

	// Read client_id and client_secret from K8s Secret
	data, err := p.reader.ReadSecret(ctx, p.secretRef.Namespace, p.secretRef.Name)
	if err != nil {
		return "", NewAuthError("MissingSecret",
			fmt.Sprintf("failed to read secret %s/%s: %v", p.secretRef.Namespace, p.secretRef.Name, err))
	}

	clientID, ok := data["clientId"]
	if !ok || len(clientID) == 0 {
		return "", NewAuthError("MissingSecret",
			fmt.Sprintf("secret %s/%s does not contain key %q", p.secretRef.Namespace, p.secretRef.Name, "clientId"))
	}

	clientSecret, ok := data["clientSecret"]
	if !ok || len(clientSecret) == 0 {
		return "", NewAuthError("MissingSecret",
			fmt.Sprintf("secret %s/%s does not contain key %q", p.secretRef.Namespace, p.secretRef.Name, "clientSecret"))
	}

	// Exchange credentials for access token
	params := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     string(clientID),
		"client_secret": string(clientSecret),
		"scope":         p.scope,
	}

	resp, err := p.exchanger.ExchangeToken(ctx, params)
	if err != nil {
		return "", NewAuthError("TokenExchangeFailure",
			fmt.Sprintf("failed to exchange client credentials: %v", err))
	}

	if resp.AccessToken == "" {
		return "", NewAuthError("TokenExchangeFailure",
			"token exchange returned empty access token")
	}

	p.token = resp.AccessToken
	if resp.ExpiresIn > 0 {
		p.expiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	} else {
		p.expiresAt = time.Now().Add(p.tokenLifetime)
	}

	return p.token, nil
}

// InvalidateToken clears the cached token so the next GetToken call performs a new exchange.
func (p *ClientCredentialsTokenProvider) InvalidateToken() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.token = ""
	p.expiresAt = time.Time{}
}
