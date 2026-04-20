// Package workspacelabel contains the Upjet resource configuration for linear_workspace_label.
package workspacelabel

import (
	"fmt"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/avodah-inc/provider-linear/internal/validation"
)

// Configure adds custom resource configuration for linear_workspace_label.
//
// External name: `label_name`.
// Validations: name (min 1), color (hex).
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_workspace_label", func(r *ujconfig.Resource) {
		r.ShortGroup = "workspacelabel"
		r.Kind = "WorkspaceLabel"
		r.Version = "v1alpha1"

		// External name: the label name serves as the import identifier.
		r.ExternalName = ujconfig.ExternalName{
			SetIdentifierArgumentFn: func(base map[string]any, externalName string) {
				base["name"] = externalName
			},
			GetExternalNameFn: func(tfstate map[string]any) (string, error) {
				if name, ok := tfstate["name"].(string); ok && name != "" {
					return name, nil
				}
				return "", fmt.Errorf("cannot determine external name: name field is empty")
			},
			OmittedFields: []string{"name"},
		}
	})
}

// ValidateWorkspaceLabelParams validates WorkspaceLabel resource parameters.
// Called by the reconciler before create and update operations.
func ValidateWorkspaceLabelParams(params map[string]any) error {
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

	return nil
}
