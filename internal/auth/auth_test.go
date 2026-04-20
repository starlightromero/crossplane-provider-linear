package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- Mock implementations ---

// mockSecretReader implements SecretReader for testing.
type mockSecretReader struct {
	data SecretData
	err  error
}

func (m *mockSecretReader) ReadSecret(_ context.Context, _, _ string) (SecretData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// mockSecretWriter implements SecretWriter for testing.
type mockSecretWriter struct {
	written   SecretData
	writeErr  error
	callCount int
}

func (m *mockSecretWriter) WriteSecret(_ context.Context, _, _ string, data SecretData) error {
	m.callCount++
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = data
	return nil
}

// mockHTTPTokenExchanger implements HTTPTokenExchanger for testing.
type mockHTTPTokenExchanger struct {
	resp      *TokenResponse
	err       error
	params    []map[string]string
	callCount int
}

func (m *mockHTTPTokenExchanger) ExchangeToken(_ context.Context, params map[string]string) (*TokenResponse, error) {
	m.callCount++
	m.params = append(m.params, params)
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

// --- SecretTokenProvider Tests ---

func TestSecretTokenProvider_Success(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{"token": []byte("lin_api_test123")},
	}
	ref := SecretRef{Namespace: "default", Name: "my-secret", Key: "token"}
	provider := NewSecretTokenProvider(reader, ref)

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "lin_api_test123" {
		t.Errorf("expected token 'lin_api_test123', got %q", token)
	}
}

func TestSecretTokenProvider_CachesToken(t *testing.T) {
	callCount := 0
	reader := &mockSecretReader{
		data: SecretData{"token": []byte("cached-token")},
	}
	// Wrap to count calls
	countingReader := &countingSecretReader{inner: reader}
	ref := SecretRef{Namespace: "default", Name: "my-secret", Key: "token"}
	provider := NewSecretTokenProvider(countingReader, ref)

	// First call reads from secret
	_, _ = provider.GetToken(context.Background())
	// Second call should use cache
	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cached-token" {
		t.Errorf("expected 'cached-token', got %q", token)
	}
	if countingReader.callCount != 1 {
		t.Errorf("expected 1 read call (cached), got %d", callCount)
	}
}

func TestSecretTokenProvider_ErrorSecretNotExist(t *testing.T) {
	reader := &mockSecretReader{
		err: errors.New("secret not found"),
	}
	ref := SecretRef{Namespace: "default", Name: "missing-secret", Key: "token"}
	provider := NewSecretTokenProvider(reader, ref)

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "MissingSecret" {
		t.Errorf("expected reason 'MissingSecret', got %q", authErr.Reason)
	}
}

func TestSecretTokenProvider_ErrorKeyMissing(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{"other-key": []byte("value")},
	}
	ref := SecretRef{Namespace: "default", Name: "my-secret", Key: "token"}
	provider := NewSecretTokenProvider(reader, ref)

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "MissingSecret" {
		t.Errorf("expected reason 'MissingSecret', got %q", authErr.Reason)
	}
}

func TestSecretTokenProvider_ErrorEmptyToken(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{"token": []byte("   ")},
	}
	ref := SecretRef{Namespace: "default", Name: "my-secret", Key: "token"}
	provider := NewSecretTokenProvider(reader, ref)

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "InvalidToken" {
		t.Errorf("expected reason 'InvalidToken', got %q", authErr.Reason)
	}
}

func TestSecretTokenProvider_InvalidateForceReread(t *testing.T) {
	reader := &countingSecretReader{
		inner: &mockSecretReader{
			data: SecretData{"token": []byte("fresh-token")},
		},
	}
	ref := SecretRef{Namespace: "default", Name: "my-secret", Key: "token"}
	provider := NewSecretTokenProvider(reader, ref)

	// First read
	_, _ = provider.GetToken(context.Background())
	if reader.callCount != 1 {
		t.Fatalf("expected 1 call, got %d", reader.callCount)
	}

	// Invalidate and re-read
	provider.InvalidateToken()
	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "fresh-token" {
		t.Errorf("expected 'fresh-token', got %q", token)
	}
	if reader.callCount != 2 {
		t.Errorf("expected 2 calls after invalidation, got %d", reader.callCount)
	}
}

func TestSecretTokenProvider_DefaultKey(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{"token": []byte("default-key-token")},
	}
	// Empty key should default to "token"
	ref := SecretRef{Namespace: "default", Name: "my-secret", Key: ""}
	provider := NewSecretTokenProvider(reader, ref)

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "default-key-token" {
		t.Errorf("expected 'default-key-token', got %q", token)
	}
}

// --- ClientCredentialsTokenProvider Tests ---

func TestClientCredentials_Success(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId":     []byte("my-client-id"),
			"clientSecret": []byte("my-client-secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{
			AccessToken: "cc-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cc-access-token" {
		t.Errorf("expected 'cc-access-token', got %q", token)
	}
	// Verify exchange params
	if len(exchanger.params) != 1 {
		t.Fatalf("expected 1 exchange call, got %d", len(exchanger.params))
	}
	if exchanger.params[0]["grant_type"] != "client_credentials" {
		t.Errorf("expected grant_type 'client_credentials', got %q", exchanger.params[0]["grant_type"])
	}
}

func TestClientCredentials_CachesToken(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId":     []byte("my-client-id"),
			"clientSecret": []byte("my-client-secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{
			AccessToken: "cached-cc-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	// First call
	_, _ = provider.GetToken(context.Background())
	// Second call should use cache
	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cached-cc-token" {
		t.Errorf("expected 'cached-cc-token', got %q", token)
	}
	if exchanger.callCount != 1 {
		t.Errorf("expected 1 exchange call (cached), got %d", exchanger.callCount)
	}
}

func TestClientCredentials_ReExchangeOnInvalidate(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId":     []byte("my-client-id"),
			"clientSecret": []byte("my-client-secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{
			AccessToken: "new-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	// First exchange
	_, _ = provider.GetToken(context.Background())
	// Invalidate (simulates 401)
	provider.InvalidateToken()
	// Should re-exchange
	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "new-token" {
		t.Errorf("expected 'new-token', got %q", token)
	}
	if exchanger.callCount != 2 {
		t.Errorf("expected 2 exchange calls, got %d", exchanger.callCount)
	}
}

func TestClientCredentials_ErrorSecretNotExist(t *testing.T) {
	reader := &mockSecretReader{
		err: errors.New("secret not found"),
	}
	exchanger := &mockHTTPTokenExchanger{}
	ref := SecretRef{Namespace: "default", Name: "missing"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "MissingSecret" {
		t.Errorf("expected reason 'MissingSecret', got %q", authErr.Reason)
	}
}

func TestClientCredentials_ErrorMissingClientId(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientSecret": []byte("secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "MissingSecret" {
		t.Errorf("expected reason 'MissingSecret', got %q", authErr.Reason)
	}
}

func TestClientCredentials_ErrorMissingClientSecret(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId": []byte("id"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "MissingSecret" {
		t.Errorf("expected reason 'MissingSecret', got %q", authErr.Reason)
	}
}

func TestClientCredentials_ErrorExchangeFails(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId":     []byte("id"),
			"clientSecret": []byte("secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{
		err: errors.New("network error"),
	}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "TokenExchangeFailure" {
		t.Errorf("expected reason 'TokenExchangeFailure', got %q", authErr.Reason)
	}
}

func TestClientCredentials_DefaultScope(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId":     []byte("id"),
			"clientSecret": []byte("secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{AccessToken: "tok", ExpiresIn: 3600},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	// Empty scope should default to "read,write"
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	_, _ = provider.GetToken(context.Background())
	if exchanger.params[0]["scope"] != "read,write" {
		t.Errorf("expected default scope 'read,write', got %q", exchanger.params[0]["scope"])
	}
}

func TestClientCredentials_CustomScope(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId":     []byte("id"),
			"clientSecret": []byte("secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{AccessToken: "tok", ExpiresIn: 3600},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "read,write,admin")

	_, _ = provider.GetToken(context.Background())
	if exchanger.params[0]["scope"] != "read,write,admin" {
		t.Errorf("expected custom scope 'read,write,admin', got %q", exchanger.params[0]["scope"])
	}
}

// --- AuthorizationCodeTokenProvider Tests ---

func TestAuthCode_SuccessLoadsTokens(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"access_token":  []byte("ac-token"),
			"refresh_token": []byte("rf-token"),
			"client_id":     []byte("cid"),
			"client_secret": []byte("csecret"),
		},
	}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}
	ref := SecretRef{Namespace: "default", Name: "oauth2-secret"}
	provider := NewAuthorizationCodeTokenProvider(reader, writer, exchanger, ref)

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ac-token" {
		t.Errorf("expected 'ac-token', got %q", token)
	}
	// No exchange should have happened (token is fresh)
	if exchanger.callCount != 0 {
		t.Errorf("expected 0 exchange calls, got %d", exchanger.callCount)
	}
}

func TestAuthCode_RefreshesExpiredToken(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"access_token":  []byte("expired-token"),
			"refresh_token": []byte("rf-token"),
			"client_id":     []byte("cid"),
			"client_secret": []byte("csecret"),
		},
	}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{
			AccessToken:  "refreshed-token",
			RefreshToken: "new-rf-token",
			ExpiresIn:    86400,
		},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth2-secret"}
	provider := NewAuthorizationCodeTokenProvider(reader, writer, exchanger, ref)

	// Load credentials first
	_, _ = provider.GetToken(context.Background())

	// Simulate expiry by invalidating
	provider.InvalidateToken()

	// Next call should trigger refresh
	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "refreshed-token" {
		t.Errorf("expected 'refreshed-token', got %q", token)
	}
	if exchanger.callCount != 1 {
		t.Errorf("expected 1 exchange call for refresh, got %d", exchanger.callCount)
	}
	// Verify refresh params
	if exchanger.params[0]["grant_type"] != "refresh_token" {
		t.Errorf("expected grant_type 'refresh_token', got %q", exchanger.params[0]["grant_type"])
	}
}

func TestAuthCode_WritesNewTokensToSecret(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"access_token":  []byte("old-token"),
			"refresh_token": []byte("old-rf"),
			"client_id":     []byte("cid"),
			"client_secret": []byte("csecret"),
		},
	}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
			ExpiresIn:    86400,
		},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth2-secret"}
	provider := NewAuthorizationCodeTokenProvider(reader, writer, exchanger, ref)

	// Load then invalidate to force refresh
	_, _ = provider.GetToken(context.Background())
	provider.InvalidateToken()
	_, _ = provider.GetToken(context.Background())

	// Verify secret was written
	if writer.callCount != 1 {
		t.Fatalf("expected 1 write call, got %d", writer.callCount)
	}
	if string(writer.written["access_token"]) != "new-access" {
		t.Errorf("expected written access_token 'new-access', got %q", string(writer.written["access_token"]))
	}
	if string(writer.written["refresh_token"]) != "new-refresh" {
		t.Errorf("expected written refresh_token 'new-refresh', got %q", string(writer.written["refresh_token"]))
	}
}

func TestAuthCode_ErrorSecretNotExist(t *testing.T) {
	reader := &mockSecretReader{
		err: errors.New("secret not found"),
	}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}
	ref := SecretRef{Namespace: "default", Name: "missing"}
	provider := NewAuthorizationCodeTokenProvider(reader, writer, exchanger, ref)

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "MissingSecret" {
		t.Errorf("expected reason 'MissingSecret', got %q", authErr.Reason)
	}
}

func TestAuthCode_ErrorMissingRequiredKeys(t *testing.T) {
	// Missing refresh_token
	reader := &mockSecretReader{
		data: SecretData{
			"access_token":  []byte("token"),
			"client_id":     []byte("cid"),
			"client_secret": []byte("csecret"),
		},
	}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}
	ref := SecretRef{Namespace: "default", Name: "incomplete-secret"}
	provider := NewAuthorizationCodeTokenProvider(reader, writer, exchanger, ref)

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "MissingSecret" {
		t.Errorf("expected reason 'MissingSecret', got %q", authErr.Reason)
	}
}

func TestAuthCode_ErrorRefreshFails(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"access_token":  []byte("old-token"),
			"refresh_token": []byte("rf"),
			"client_id":     []byte("cid"),
			"client_secret": []byte("csecret"),
		},
	}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{
		err: errors.New("refresh failed"),
	}
	ref := SecretRef{Namespace: "default", Name: "oauth2-secret"}
	provider := NewAuthorizationCodeTokenProvider(reader, writer, exchanger, ref)

	// Load credentials
	_, _ = provider.GetToken(context.Background())
	// Force expiry
	provider.InvalidateToken()

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Reason != "RefreshFailure" {
		t.Errorf("expected reason 'RefreshFailure', got %q", authErr.Reason)
	}
}

func TestAuthCode_InvalidateForceRefresh(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"access_token":  []byte("original"),
			"refresh_token": []byte("rf"),
			"client_id":     []byte("cid"),
			"client_secret": []byte("csecret"),
		},
	}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{
			AccessToken:  "after-invalidate",
			RefreshToken: "new-rf",
			ExpiresIn:    86400,
		},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth2-secret"}
	provider := NewAuthorizationCodeTokenProvider(reader, writer, exchanger, ref)

	// Load
	token1, _ := provider.GetToken(context.Background())
	if token1 != "original" {
		t.Errorf("expected 'original', got %q", token1)
	}

	// Invalidate simulates 401
	provider.InvalidateToken()

	// Should refresh
	token2, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token2 != "after-invalidate" {
		t.Errorf("expected 'after-invalidate', got %q", token2)
	}
}

// --- Validation Tests ---

func TestValidateCredentialSource_ValidSources(t *testing.T) {
	validSources := []CredentialSource{SourceSecret, SourceOAuth2ClientCredentials, SourceOAuth2}
	for _, src := range validSources {
		if err := ValidateCredentialSource(src); err != nil {
			t.Errorf("ValidateCredentialSource(%q) unexpected error: %v", src, err)
		}
	}
}

func TestValidateCredentialSource_InvalidSources(t *testing.T) {
	invalidSources := []CredentialSource{"", "Invalid", "BasicAuth", "oauth2"}
	for _, src := range invalidSources {
		if err := ValidateCredentialSource(src); err == nil {
			t.Errorf("ValidateCredentialSource(%q) expected error, got nil", src)
		}
	}
}

func TestValidateCredentials_NilCredentials(t *testing.T) {
	err := ValidateCredentials(nil)
	if err == nil {
		t.Fatal("expected error for nil credentials, got nil")
	}
}

func TestValidateCredentials_MissingSecretRef(t *testing.T) {
	creds := &Credentials{
		Source:    SourceSecret,
		SecretRef: SecretRef{Namespace: "default", Name: ""},
	}
	err := ValidateCredentials(creds)
	if err == nil {
		t.Fatal("expected error for missing secretRef.name, got nil")
	}
}

func TestValidateCredentials_MissingNamespace(t *testing.T) {
	creds := &Credentials{
		Source:    SourceSecret,
		SecretRef: SecretRef{Namespace: "", Name: "my-secret"},
	}
	err := ValidateCredentials(creds)
	if err == nil {
		t.Fatal("expected error for missing secretRef.namespace, got nil")
	}
}

func TestValidateCredentials_ValidSecret(t *testing.T) {
	creds := &Credentials{
		Source:    SourceSecret,
		SecretRef: SecretRef{Namespace: "default", Name: "my-secret", Key: "token"},
	}
	err := ValidateCredentials(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewTokenProvider_CreatesSecretProvider(t *testing.T) {
	creds := &Credentials{
		Source:    SourceSecret,
		SecretRef: SecretRef{Namespace: "ns", Name: "sec", Key: "tok"},
	}
	reader := &mockSecretReader{}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}

	provider, err := NewTokenProvider(creds, reader, writer, exchanger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := provider.(*SecretTokenProvider); !ok {
		t.Errorf("expected *SecretTokenProvider, got %T", provider)
	}
}

func TestNewTokenProvider_CreatesClientCredentialsProvider(t *testing.T) {
	creds := &Credentials{
		Source:    SourceOAuth2ClientCredentials,
		SecretRef: SecretRef{Namespace: "ns", Name: "sec"},
	}
	reader := &mockSecretReader{}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}

	provider, err := NewTokenProvider(creds, reader, writer, exchanger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := provider.(*ClientCredentialsTokenProvider); !ok {
		t.Errorf("expected *ClientCredentialsTokenProvider, got %T", provider)
	}
}

func TestNewTokenProvider_CreatesAuthCodeProvider(t *testing.T) {
	creds := &Credentials{
		Source:    SourceOAuth2,
		SecretRef: SecretRef{Namespace: "ns", Name: "sec"},
	}
	reader := &mockSecretReader{}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}

	provider, err := NewTokenProvider(creds, reader, writer, exchanger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := provider.(*AuthorizationCodeTokenProvider); !ok {
		t.Errorf("expected *AuthorizationCodeTokenProvider, got %T", provider)
	}
}

func TestNewTokenProvider_RejectsInvalidSource(t *testing.T) {
	creds := &Credentials{
		Source:    "UnsupportedSource",
		SecretRef: SecretRef{Namespace: "ns", Name: "sec"},
	}
	reader := &mockSecretReader{}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}

	_, err := NewTokenProvider(creds, reader, writer, exchanger)
	if err == nil {
		t.Fatal("expected error for unsupported source, got nil")
	}
}

func TestNewTokenProvider_RejectsNilCredentials(t *testing.T) {
	reader := &mockSecretReader{}
	writer := &mockSecretWriter{}
	exchanger := &mockHTTPTokenExchanger{}

	_, err := NewTokenProvider(nil, reader, writer, exchanger)
	if err == nil {
		t.Fatal("expected error for nil credentials, got nil")
	}
}

// --- Helper types ---

// countingSecretReader wraps a SecretReader and counts calls.
type countingSecretReader struct {
	inner     SecretReader
	callCount int
}

func (c *countingSecretReader) ReadSecret(ctx context.Context, namespace, name string) (SecretData, error) {
	c.callCount++
	return c.inner.ReadSecret(ctx, namespace, name)
}

// --- AuthError Tests ---

func TestAuthError_ErrorString(t *testing.T) {
	err := NewAuthError("MissingSecret", "secret not found")
	expected := "MissingSecret: secret not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

// --- Token Expiry Tests ---

func TestClientCredentials_ExpiryFromResponse(t *testing.T) {
	reader := &mockSecretReader{
		data: SecretData{
			"clientId":     []byte("id"),
			"clientSecret": []byte("secret"),
		},
	}
	exchanger := &mockHTTPTokenExchanger{
		resp: &TokenResponse{
			AccessToken: "short-lived",
			ExpiresIn:   1, // 1 second
		},
	}
	ref := SecretRef{Namespace: "default", Name: "oauth-secret"}
	provider := NewClientCredentialsTokenProvider(reader, exchanger, ref, "")

	_, _ = provider.GetToken(context.Background())
	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Should re-exchange since token expired
	exchanger.resp = &TokenResponse{AccessToken: "renewed", ExpiresIn: 3600}
	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "renewed" {
		t.Errorf("expected 'renewed', got %q", token)
	}
	if exchanger.callCount != 2 {
		t.Errorf("expected 2 exchange calls, got %d", exchanger.callCount)
	}
}
