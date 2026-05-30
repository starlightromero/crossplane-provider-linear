// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/avodah-inc/provider-linear/apis/v1alpha1"
	"github.com/avodah-inc/provider-linear/internal/auth"
)

const keyToken = "token"

// TerraformSetupBuilder returns a terraform.SetupFn that resolves credentials
// from the ProviderConfig and configures the Terraform Linear provider.
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, c client.Client, mg resource.Managed) (terraform.Setup, error) {
		setup := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		pcr, ok := mg.(resource.TypedProviderConfigReferencer)
		if !ok {
			return setup, errors.New("managed resource does not implement TypedProviderConfigReferencer")
		}
		configRef := pcr.GetProviderConfigReference()
		if configRef == nil {
			return setup, errors.New("no providerConfigRef set on managed resource")
		}

		pc := &v1alpha1.ProviderConfig{}
		if err := c.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return setup, errors.Wrap(err, "cannot get ProviderConfig")
		}

		token, err := resolveToken(ctx, c, pc)
		if err != nil {
			return setup, errors.Wrap(err, "cannot resolve credentials")
		}

		setup.Configuration = terraform.ProviderConfiguration{
			keyToken: token,
		}
		return setup, nil
	}
}

// resolveToken extracts or exchanges credentials based on the source type.
func resolveToken(ctx context.Context, c client.Client, pc *v1alpha1.ProviderConfig) (string, error) {
	ref := pc.Spec.Credentials.CommonCredentialSelectors.SecretRef
	if ref == nil {
		return "", errors.New("spec.credentials.secretRef is required")
	}

	secret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}, secret); err != nil {
		return "", errors.Wrapf(err, "cannot get secret %s/%s", ref.Namespace, ref.Name)
	}

	switch auth.CredentialSource(pc.Spec.Credentials.Source) {
	case auth.SourceOAuth2ClientCredentials:
		return exchangeClientCredentials(ctx, secret, pc.Spec.Credentials.Scope)
	case auth.SourceSecret:
		return extractSecretToken(secret, ref.Key)
	default:
		return extractSecretToken(secret, ref.Key)
	}
}

// extractSecretToken reads a token directly from a secret key.
func extractSecretToken(secret *corev1.Secret, key string) (string, error) {
	if key == "" {
		key = "token"
	}
	token := strings.TrimSpace(string(secret.Data[key]))
	if token == "" {
		return "", fmt.Errorf("secret %s/%s key %q is empty", secret.Namespace, secret.Name, key)
	}
	return token, nil
}

// exchangeClientCredentials performs the OAuth2 client credentials token exchange
// against Linear's token endpoint.
func exchangeClientCredentials(ctx context.Context, secret *corev1.Secret, scope string) (string, error) {
	clientID := strings.TrimSpace(string(secret.Data["clientId"]))
	clientSecret := strings.TrimSpace(string(secret.Data["clientSecret"]))

	if clientID == "" {
		return "", errors.New("secret missing 'clientId' key")
	}
	if clientSecret == "" {
		return "", errors.New("secret missing 'clientSecret' key")
	}
	if scope == "" {
		scope = auth.DefaultScope
	}

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"scope":         {scope},
	}

	// Use a fresh context for the HTTP request to avoid inheriting the
	// reconciler's short deadline which causes premature cancellation.
	httpCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, http.MethodPost, auth.LinearTokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "cannot create token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "token exchange request failed")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp auth.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", errors.Wrap(err, "cannot parse token response")
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("token exchange returned empty access_token")
	}

	return tokenResp.AccessToken, nil
}
