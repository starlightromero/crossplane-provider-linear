// Package teammembership contains the Upjet resource configuration for linear_team_membership.
package teammembership

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

const (
	extractorPackagePath = "github.com/crossplane/upjet/v2/pkg/resource"
)

// Configure adds custom resource configuration for linear_team_membership.
//
// External name: Linear-assigned UUID (membership ID).
// Immutable: team_id, user_id (cannot be changed after creation).
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_team_membership", func(r *ujconfig.Resource) {
		r.ShortGroup = ""
		r.Kind = "TeamMembership"
		r.Version = "v1alpha1"

		// Cross-resource reference: team_id → linear_team.
		r.References["team_id"] = ujconfig.Reference{
			TerraformName: "linear_team",
			Extractor:     extractorPackagePath + ".ExtractParamPath(\"id\", true)",
		}

		// Mark team_id and user_id as immutable.
		r.LateInitializer = ujconfig.LateInitializer{
			IgnoredFields: []string{"team_id", "user_id"},
		}
	})
}
