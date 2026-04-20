// Package workflowstate contains the Upjet resource configuration for linear_workflow_state.
package workflowstate

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

// Configure adds custom resource configuration for linear_workflow_state.
//
// External name: composite pattern `workflow_state_name:team_key`.
// Validations: name (min 1), type (enum), color (hex), teamId (UUID).
// Immutable: type and teamId (cannot be changed after creation).
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_workflow_state", func(r *ujconfig.Resource) {
		r.ShortGroup = "workflowstate"
		r.Kind = "WorkflowState"
		r.Version = "v1alpha1"

		// External name: composite key workflow_state_name:team_key.
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
			GetIDFn: func(_ string, _ string, externalName string) (string, error) {
				return externalName, nil
			},
			OmittedFields: []string{"name", "team_key"},
		}

		// Cross-resource reference: team_id → linear_team.
		// Enables teamIdRef / teamIdSelector fields on the WorkflowState CRD
		// to resolve the team UUID from a Team managed resource.
		r.References["team_id"] = ujconfig.Reference{
			TerraformName: "linear_team",
			Extractor:     extractorPackagePath + ".ExtractParamPath(\"id\", true)",
		}

		// Mark type and teamId as immutable: exclude from late
		// initialization so that updates cannot change these fields.
		r.LateInitializer = ujconfig.LateInitializer{
			IgnoredFields: []string{"type", "team_id"},
		}
	})
}

// ValidateWorkflowStateParams validates WorkflowState resource parameters.
// Called by the reconciler before create and update operations.
func ValidateWorkflowStateParams(params map[string]any) error {
	if name, ok := params["name"].(string); ok {
		if err := validation.ValidateName("name", name); err != nil {
			return err
		}
	}

	if stateType, ok := params["type"].(string); ok && stateType != "" {
		if err := validation.ValidateWorkflowStateType("type", stateType); err != nil {
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

// CheckImmutableFields verifies that type and teamId have not changed between
// the observed state and the desired state. Returns an error if either field
// was modified after creation.
func CheckImmutableFields(observed, desired map[string]any) error {
	if err := checkFieldImmutable(observed, desired, "type"); err != nil {
		return err
	}
	return checkFieldImmutable(observed, desired, "team_id")
}

func checkFieldImmutable(observed, desired map[string]any, field string) error {
	obs, _ := observed[field].(string)
	des, _ := desired[field].(string)
	if obs != "" && des != "" && obs != des {
		return fmt.Errorf("%s is immutable after creation: cannot change from %q to %q", field, obs, des)
	}
	return nil
}
