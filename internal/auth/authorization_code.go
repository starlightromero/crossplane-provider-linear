package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// AuthorizationCodeTokenProvider implements TokenProvider using OAuth2 authorization code flow
// with automatic token refresh.
type AuthorizationCodeTokenProvider struct {
	mu            sync.Mutex
	reader        SecretReader
	writer        SecretWriter
	exchanger     HTTPTokenExchanger
	secretRef     SecretRef
	accessToken   string
	refreshToken  string
	clientID      string
	clientSecret  string
	expiresAt     time.Time
	tokenLifetime time.Duration
}

// NewAuthorizationCodeTokenProvider creates a TokenProvider that uses pre-obtained OAuth2 tokens
// with automatic refresh on expiry.
func NewAuthorizationCodeTokenProvider(reader SecretReader, writer SecretWriter, exchanger HTTPTokenExchanger, ref SecretRef) *AuthorizationCodeTokenProvider {
	return &AuthorizationCodeTokenProvider{
		reader:        reader,
		writer:        writer,
		exchanger:     exchanger,
		secretRef:     ref,
		tokenLifetime: DefaultAuthCodeLifetime,
	}
}

// GetToken returns the current access token, refreshing it if expired.
func (p *AuthorizationCodeTokenProvider) GetToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If we have a valid cached token, return it
	if p.accessToken != "" && time.Now().Before(p.expiresAt) {
		return p.accessToken, nil
	}

	// Load credentials from Secret if not yet loaded
	if p.clientID == "" {
		if err := p.loadCredentials(ctx); err != nil {
			return "", err
		}
	}

	// If we have an access token but it's expired, refresh it
	if p.accessToken != "" && !time.Now().Before(p.expiresAt) {
		return p.refresh(ctx)
	}

	// First-time load: use the access token from the secret
	if p.accessToken != "" {
		return p.accessToken, nil
	}

	return "", NewAuthError("MissingToken",
		"no access token available and no refresh token to obtain one")
}

// InvalidateToken marks the current token as invalid, forcing a refresh on next GetToken.
func (p *AuthorizationCodeTokenProvider) InvalidateToken() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.expiresAt = time.Time{} // Force refresh on next call
}

// loadCredentials reads all required OAuth2 fields from the K8s Secret.
func (p *AuthorizationCodeTokenProvider) loadCredentials(ctx context.Context) error {
	data, err := p.reader.ReadSecret(ctx, p.secretRef.Namespace, p.secretRef.Name)
	if err != nil {
		return NewAuthError("MissingSecret",
			fmt.Sprintf("failed to read secret %s/%s: %v", p.secretRef.Namespace, p.secretRef.Name, err))
	}

	requiredKeys := []string{"access_token", "refresh_token", "client_id", "client_secret"}
	values := make(map[string]string, len(requiredKeys))
	for _, key := range requiredKeys {
		v := strings.TrimSpace(string(data[key]))
		if v == "" {
			return NewAuthError("MissingSecret",
				fmt.Sprintf("secret %s/%s is missing required key %q", p.secretRef.Namespace, p.secretRef.Name, key))
		}
		values[key] = v
	}

	p.accessToken = values["access_token"]
	p.refreshToken = values["refresh_token"]
	p.clientID = values["client_id"]
	p.clientSecret = values["client_secret"]
	// Assume the token is valid for the full lifetime from now (conservative)
	p.expiresAt = time.Now().Add(p.tokenLifetime)

	return nil
}

// refresh exchanges the refresh token for a new access token and writes the new tokens
// back to the K8s Secret.
func (p *AuthorizationCodeTokenProvider) refresh(ctx context.Context) (string, error) {
	if p.refreshToken == "" {
		return "", NewAuthError("RefreshFailure",
			"no refresh token available; re-authorization required")
	}

	params := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": p.refreshToken,
		"client_id":     p.clientID,
		"client_secret": p.clientSecret,
	}

	resp, err := p.exchanger.ExchangeToken(ctx, params)
	if err != nil {
		return "", NewAuthError("RefreshFailure",
			fmt.Sprintf("token refresh failed: %v; re-authorization required", err))
	}

	if resp.AccessToken == "" {
		return "", NewAuthError("RefreshFailure",
			"token refresh returned empty access token; re-authorization required")
	}

	// Update cached tokens
	p.accessToken = resp.AccessToken
	if resp.RefreshToken != "" {
		p.refreshToken = resp.RefreshToken
	}
	if resp.ExpiresIn > 0 {
		p.expiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	} else {
		p.expiresAt = time.Now().Add(p.tokenLifetime)
	}

	// Write new tokens back to K8s Secret
	secretData := SecretData{
		"access_token":  []byte(p.accessToken),
		"refresh_token": []byte(p.refreshToken),
		"client_id":     []byte(p.clientID),
		"client_secret": []byte(p.clientSecret),
	}

	if err := p.writer.WriteSecret(ctx, p.secretRef.Namespace, p.secretRef.Name, secretData); err != nil {
		// Token refresh succeeded but secret write failed — log but don't fail
		// The token is still valid in memory
		return p.accessToken, nil
	}

	return p.accessToken, nil
}
