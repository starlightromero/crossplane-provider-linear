// Package config contains the Upjet provider configuration for provider-linear.
package config

import "github.com/crossplane/upjet/v2/pkg/config"

// ExternalNameConfigs contains all external name configurations for this
// provider. Each entry maps a Terraform resource name to its Crossplane
// external name configuration, defining how Terraform import keys translate
// to Crossplane external name annotations.
//
// Resource → External Name Pattern:
//
//	linear_team              → key (e.g., "ENG")
//	linear_team_label        → label_name:team_key (composite)
//	linear_team_workflow     → team_key or team_key:branch_pattern:is_regex
//	linear_template          → id (Linear-assigned UUID)
//	linear_workflow_state    → workflow_state_name:team_key (composite)
//	linear_workspace_label   → label_name
//	linear_workspace_settings → id (Linear-assigned UUID)
//	linear_workspace         → id (Linear-assigned UUID, data source)
var ExternalNameConfigs = map[string]config.ExternalName{
	// Team: imported by team key (e.g., "ENG"). The key field serves as the
	// Terraform import identifier.
	"linear_team": config.IdentifierFromProvider,

	// TeamLabel: imported by composite key label_name:team_key.
	// Terraform assigns the identifier on creation.
	"linear_team_label": config.IdentifierFromProvider,

	// TeamWorkflow: imported by team_key or team_key:branch_pattern:is_regex.
	// Terraform assigns the identifier on creation.
	"linear_team_workflow": config.IdentifierFromProvider,

	// Template: imported by Linear-assigned UUID.
	"linear_template": config.IdentifierFromProvider,

	// WorkflowState: imported by composite key workflow_state_name:team_key.
	// Terraform assigns the identifier on creation.
	"linear_workflow_state": config.IdentifierFromProvider,

	// WorkspaceLabel: imported by label_name.
	// Terraform assigns the identifier on creation.
	"linear_workspace_label": config.IdentifierFromProvider,

	// WorkspaceSettings: imported by Linear-assigned UUID (singleton).
	"linear_workspace_settings": config.IdentifierFromProvider,

	// Workspace: data source, imported by Linear-assigned UUID.
	"linear_workspace": config.IdentifierFromProvider,
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs. This is passed as a default resource option to
// the provider so that each resource gets its external name configuration
// from the map.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually. Each entry is suffixed with "$" for regex matching
// in the Upjet include list.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
