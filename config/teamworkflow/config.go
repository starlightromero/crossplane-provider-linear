// Package teamworkflow contains the Upjet resource configuration for linear_team_workflow.
package teamworkflow

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

// Configure adds custom resource configuration for linear_team_workflow.
//
// External name: `team_key` or `team_key:branch_pattern:is_regex`.
// Validations: workflow state UUID fields (draft, start, review, mergeable, merge).
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_team_workflow", func(r *ujconfig.Resource) {
		r.ShortGroup = "teamworkflow"
		r.Kind = "TeamWorkflow"
		r.Version = "v1alpha1"

		// External name: team_key or team_key:branch_pattern:is_regex.
		r.ExternalName = ujconfig.ExternalName{
			SetIdentifierArgumentFn: func(base map[string]any, externalName string) {
				parts := strings.SplitN(externalName, ":", 3)
				if len(parts) >= 1 {
					base["key"] = parts[0]
				}
				if len(parts) >= 3 {
					base["branch_pattern"] = parts[1]
					base["is_regex"] = parts[2]
				}
			},
			GetExternalNameFn: func(tfstate map[string]any) (string, error) {
				key, _ := tfstate["key"].(string)
				if key == "" {
					return "", fmt.Errorf("cannot determine external name: key field is empty")
				}
				pattern, _ := tfstate["branch_pattern"].(string)
				isRegex, _ := tfstate["is_regex"].(string)
				if pattern != "" {
					return key + ":" + pattern + ":" + isRegex, nil
				}
				return key, nil
			},
			GetIDFn: func(_ string, _ string, externalName string) (string, error) {
				return externalName, nil
			},
			OmittedFields: []string{"key"},
		}

		// Cross-resource references: workflow state fields → linear_workflow_state.
		// Each field (draft, start, review, mergeable, merge) can reference a
		// WorkflowState managed resource via ref/selector fields.
		for _, field := range []string{"draft", "start", "review", "mergeable", "merge"} {
			r.References[field] = ujconfig.Reference{
				TerraformName: "linear_workflow_state",
				Extractor:     extractorPackagePath + ".ExtractParamPath(\"id\", true)",
			}
		}
	})
}

// workflowStateUUIDFields lists the TeamWorkflow fields that must be valid UUIDs.
var workflowStateUUIDFields = []string{"draft", "start", "review", "mergeable", "merge"}

// ValidateTeamWorkflowParams validates TeamWorkflow resource parameters.
// Called by the reconciler before create and update operations.
func ValidateTeamWorkflowParams(params map[string]any) error {
	for _, field := range workflowStateUUIDFields {
		if v, ok := params[field].(string); ok && v != "" {
			if err := validation.ValidateUUID(field, v); err != nil {
				return err
			}
		}
	}
	return nil
}
