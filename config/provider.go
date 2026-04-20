// Package config contains the Upjet provider configuration for provider-linear.
//
// It maps each upstream Terraform resource from terraform-community-providers/linear
// to a Crossplane managed resource under the CRD group linear.crossplane.io.
package config

import (
	// Embed provider schema and metadata for Upjet code generation.
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/avodah-inc/provider-linear/config/team"
	"github.com/avodah-inc/provider-linear/config/teamlabel"
	"github.com/avodah-inc/provider-linear/config/teamworkflow"
	"github.com/avodah-inc/provider-linear/config/template"
	"github.com/avodah-inc/provider-linear/config/workflowstate"
	"github.com/avodah-inc/provider-linear/config/workspace"
	"github.com/avodah-inc/provider-linear/config/workspacelabel"
	"github.com/avodah-inc/provider-linear/config/workspacesettings"
)

const (
	modulePath = "github.com/avodah-inc/provider-linear"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns the Upjet provider configuration for provider-linear.
//
// It registers the upstream Terraform provider schema and applies per-resource
// configuration (external names, field validations, cross-resource references,
// immutability guards) for each of the seven managed resources and the
// Workspace data source.
//
// All CRDs are registered under the group linear.crossplane.io with API
// version v1alpha1.
//
// Cross-resource reference resolution behavior (Requirement 14.5):
// Crossplane's built-in reference resolution mechanism automatically sets
// the Synced condition to False when a referenced resource does not exist
// or is not yet ready. The reconciler will not proceed with create/update
// operations until all references are resolved. This is handled by the
// Upjet framework's reference resolution pipeline — no custom logic is
// needed beyond declaring r.References in each resource configurator.
func GetProvider() (*ujconfig.Provider, error) {
	pc := ujconfig.NewProvider(
		[]byte(providerSchema),
		ProviderShortName,
		modulePath,
		[]byte(providerMetadata),
		ujconfig.WithRootGroup(ProviderCRDGroup),
		ujconfig.WithShortName(ProviderShortName),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		),
	)

	// Register per-resource configurations.
	for _, configure := range []func(provider *ujconfig.Provider){
		team.Configure,
		teamlabel.Configure,
		teamworkflow.Configure,
		template.Configure,
		workflowstate.Configure,
		workspacelabel.Configure,
		workspacesettings.Configure,
		workspace.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc, nil
}
