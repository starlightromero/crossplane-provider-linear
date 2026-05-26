// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package clients

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pcv1alpha1 "github.com/avodah-inc/provider-linear/apis/v1alpha1"
	"github.com/avodah-inc/provider-linear/internal/auth"
)

// ResolveTokenFromProviderConfig resolves a Linear API token from a ProviderConfig name.
func ResolveTokenFromProviderConfig(ctx context.Context, kube client.Client, pcName string) (string, error) {
	pc := &pcv1alpha1.ProviderConfig{}
	if err := kube.Get(ctx, types.NamespacedName{Name: pcName}, pc); err != nil {
		return "", errors.Wrap(err, "cannot get ProviderConfig")
	}

	ref := pc.Spec.Credentials.CommonCredentialSelectors.SecretRef
	if ref == nil {
		return "", errors.New("spec.credentials.secretRef is required")
	}

	secret := &corev1.Secret{}
	if err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, secret); err != nil {
		return "", errors.Wrap(err, "cannot get credentials secret")
	}

	if auth.CredentialSource(pc.Spec.Credentials.Source) == auth.SourceOAuth2ClientCredentials {
		return exchangeClientCredentials(ctx, secret, pc.Spec.Credentials.Scope)
	}

	key := ref.Key
	if key == "" {
		key = "token"
	}
	token := strings.TrimSpace(string(secret.Data[key]))
	if token == "" {
		return "", errors.Errorf("secret key %q is empty", key)
	}
	return token, nil
}
