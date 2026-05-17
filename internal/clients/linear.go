// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package clients

import (
	"context"
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/avodah-inc/provider-linear/apis/v1alpha1"
)

const keyToken = "token"

// TerraformSetupBuilder returns a terraform.SetupFn that reads credentials
// from the ProviderConfig's referenced Secret and configures the Terraform
// Linear provider.
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		setup := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		pcr, ok := mg.(interface{ GetProviderConfigReference() *xpv1.Reference })
		if !ok || pcr.GetProviderConfigReference() == nil {
			return setup, errors.New("no providerConfigRef set on managed resource")
		}
		configRef := pcr.GetProviderConfigReference()

		pc := &v1alpha1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return setup, errors.Wrap(err, "cannot get ProviderConfig")
		}

		creds, err := resource.CommonCredentialExtractor(
			ctx,
			pc.Spec.Credentials.Source,
			client,
			pc.Spec.Credentials.CommonCredentialSelectors,
		)
		if err != nil {
			return setup, errors.Wrap(err, "cannot extract credentials")
		}

		token := strings.TrimSpace(string(creds))
		if token == "" {
			return setup, errors.New("credentials are empty")
		}

		setup.Configuration = terraform.ProviderConfiguration{
			keyToken: token,
		}

		return setup, nil
	}
}
