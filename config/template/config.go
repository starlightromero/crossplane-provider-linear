// Package template contains the Upjet resource configuration for linear_template.
package template

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/avodah-inc/provider-linear/internal/validation"
)

const (
	// extractorPackagePath is the import path for the Upjet resource extractor utilities.
	extractorPackagePath = "github.com/crossplane/upjet/v2/pkg/resource"
)

// DefaultTemplateType is the default value for the Template type field when
// omitted by the user.
const DefaultTemplateType = "issue"

// Configure adds custom resource configuration for linear_template.
//
// External name: Linear-assigned UUID (`id`).
// Validations: name (min 1), data (valid JSON), type (enum, default "issue").
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_template", func(r *ujconfig.Resource) {
		r.ShortGroup = "template"
		r.Kind = "Template"
		r.Version = "v1alpha1"

		// External name: Linear-assigned UUID. The provider assigns the
		// identifier on creation; Crossplane reads it back from state.
		r.ExternalName = ujconfig.ExternalName{
			SetIdentifierArgumentFn: func(base map[string]any, externalName string) {
				base["id"] = externalName
			},
			GetExternalNameFn: func(tfstate map[string]any) (string, error) {
				if id, ok := tfstate["id"].(string); ok && id != "" {
					return id, nil
				}
				return "", nil
			},
			GetIDFn: func(_ string, _ string, externalName string) (string, error) {
				return externalName, nil
			},
			OmittedFields: []string{"id"},
		}

		// Cross-resource reference: team_id → linear_team.
		// Enables teamIdRef / teamIdSelector fields on the Template CRD
		// to resolve the team UUID from a Team managed resource.
		// When teamId is omitted, the template is workspace-scoped.
		r.References["team_id"] = ujconfig.Reference{
			TerraformName: "linear_team",
			Extractor:     extractorPackagePath + ".ExtractParamPath(\"id\", true)",
		}
	})
}

// ValidateTemplateParams validates Template resource parameters. Called by
// the reconciler before create and update operations.
func ValidateTemplateParams(params map[string]any) error {
	if name, ok := params["name"].(string); ok {
		if err := validation.ValidateName("name", name); err != nil {
			return err
		}
	}

	if data, ok := params["data"].(string); ok && data != "" {
		if err := validation.ValidateJSON("data", data); err != nil {
			return err
		}
	}

	// Apply default type if not specified.
	templateType, ok := params["type"].(string)
	if !ok || templateType == "" {
		templateType = DefaultTemplateType
	}
	if err := validation.ValidateTemplateType("type", templateType); err != nil {
		return err
	}

	return nil
}
