// Package workspace contains the Upjet data source configuration for linear_workspace.
package workspace

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure adds custom data source configuration for linear_workspace.
//
// External name: Linear-assigned UUID (`id`).
// Read-only: no writable spec fields; status populates id, name, urlKey.
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_workspace", func(r *ujconfig.Resource) {
		r.ShortGroup = "workspace"
		r.Kind = "Workspace"
		r.Version = "v1alpha1"

		// External name: Linear-assigned UUID. The workspace is a
		// read-only data source; the provider reads the identifier.
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
			OmittedFields: []string{"id"},
		}

		// Read-only data source: no writable spec fields. Upjet maps
		// the Terraform data source attributes to status.atProvider:
		//   id      → status.atProvider.id
		//   name    → status.atProvider.name
		//   url_key → status.atProvider.urlKey
		//
		// The forProvider section is empty since this is a data source
		// with no configurable inputs.
	})
}
