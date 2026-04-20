// Package teamlabel contains the Upjet resource configuration for linear_team_label.
package teamlabel

import (
	"fmt"
	"strings"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/avodah-inc/provider-linear/internal/validation"
)

const (
	// extractorPackagePath is the import path for the Upjet resource extractor utilities.
	extractorPackagePath = "github.com/crossplane/upjet/v2/pkg/resource"
)

// Configure adds custom resource configuration for linear_team_label.
//
// External name: composite pattern `label_name:team_key`.
// Validations: name (min 1), color (hex), teamId (UUID).
// Immutable: teamId (cannot be changed after creation).
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_team_label", func(r *ujconfig.Resource) {
		r.ShortGroup = "teamlabel"
		r.Kind = "TeamLabel"
		r.Version = "v1alpha1"

		// External name: composite key label_name:team_key.
		r.ExternalName = ujconfig.ExternalName{
			SetIdentifierArgumentFn: func(base map[string]any, externalName string) {
				parts := strings.SplitN(externalName, ":", 2)
				if len(parts) == 2 {
					base["name"] = parts[0]
					base["team_key"] = parts[1]
				}
			},
			GetExternalNameFn: func(tfstate map[string]any) (string, error) {
				name, _ := tfstate["name"].(string)
				teamKey, _ := tfstate["team_key"].(string)
				if name == "" || teamKey == "" {
					return "", fmt.Errorf("cannot determine external name: name or team_key is empty")
				}
				return name + ":" + teamKey, nil
			},
			OmittedFields: []string{"name", "team_key"},
		}

		// Cross-resource reference: team_id → linear_team.
		// Enables teamIdRef / teamIdSelector fields on the TeamLabel CRD
		// to resolve the team UUID from a Team managed resource.
		r.References["team_id"] = ujconfig.Reference{
			TerraformName: "linear_team",
			Extractor:     extractorPackagePath + ".ExtractParamPath(\"id\", true)",
		}

		// Mark teamId as immutable: exclude from late initialization so
		// that updates cannot change the team association.
		r.LateInitializer = ujconfig.LateInitializer{
			IgnoredFields: []string{"team_id"},
		}
	})
}

// ValidateTeamLabelParams validates TeamLabel resource parameters. Called by
// the reconciler before create and update operations.
func ValidateTeamLabelParams(params map[string]any) error {
	if name, ok := params["name"].(string); ok {
		if err := validation.ValidateName("name", name); err != nil {
			return err
		}
	}

	if color, ok := params["color"].(string); ok && color != "" {
		if err := validation.ValidateHexColor("color", color); err != nil {
			return err
		}
	}

	if teamID, ok := params["team_id"].(string); ok && teamID != "" {
		if err := validation.ValidateUUID("teamId", teamID); err != nil {
			return err
		}
	}

	return nil
}

// CheckTeamIDImmutable verifies that teamId has not changed between the
// observed state and the desired state. Returns an error if the field was
// modified after creation.
func CheckTeamIDImmutable(observed, desired map[string]any) error {
	obsTeamID, _ := observed["team_id"].(string)
	desTeamID, _ := desired["team_id"].(string)
	if obsTeamID != "" && desTeamID != "" && obsTeamID != desTeamID {
		return fmt.Errorf("teamId is immutable after creation: cannot change from %q to %q", obsTeamID, desTeamID)
	}
	return nil
}
