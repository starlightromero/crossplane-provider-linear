package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// SecretTokenProvider implements TokenProvider using a static API token from a K8s Secret.
type SecretTokenProvider struct {
	mu        sync.Mutex
	reader    SecretReader
	secretRef SecretRef
	token     string
}

// NewSecretTokenProvider creates a TokenProvider that reads an API token from a K8s Secret.
func NewSecretTokenProvider(reader SecretReader, ref SecretRef) *SecretTokenProvider {
	return &SecretTokenProvider{
		reader:    reader,
		secretRef: ref,
	}
}

// GetToken reads the API token from the referenced K8s Secret and returns it as a Bearer token.
func (p *SecretTokenProvider) GetToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token != "" {
		return p.token, nil
	}

	data, err := p.reader.ReadSecret(ctx, p.secretRef.Namespace, p.secretRef.Name)
	if err != nil {
		return "", NewAuthError("MissingSecret",
			fmt.Sprintf("failed to read secret %s/%s: %v", p.secretRef.Namespace, p.secretRef.Name, err))
	}

	key := p.secretRef.Key
	if key == "" {
		key = "token"
	}

	tokenBytes, ok := data[key]
	if !ok {
		return "", NewAuthError("MissingSecret",
			fmt.Sprintf("secret %s/%s does not contain key %q", p.secretRef.Namespace, p.secretRef.Name, key))
	}

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", NewAuthError("InvalidToken",
			fmt.Sprintf("secret %s/%s key %q is empty", p.secretRef.Namespace, p.secretRef.Name, key))
	}

	p.token = token
	return p.token, nil
}

// InvalidateToken clears the cached token so the next GetToken call re-reads from the Secret.
func (p *SecretTokenProvider) InvalidateToken() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.token = ""
}
