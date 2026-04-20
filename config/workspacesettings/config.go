// Package workspacesettings contains the Upjet resource configuration for
// linear_workspace_settings.
package workspacesettings

import (
	"fmt"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/avodah-inc/provider-linear/internal/validation"
)

// Configure adds custom resource configuration for linear_workspace_settings.
//
// External name: Linear-assigned UUID (`id`).
// Validations: fiscalYearStartMonth (range 0-11).
// Singleton: only one WorkspaceSettings resource per workspace.
// Sub-objects: projects, initiatives, feed, customers.
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_workspace_settings", func(r *ujconfig.Resource) {
		r.ShortGroup = "workspacesettings"
		r.Kind = "WorkspaceSettings"
		r.Version = "v1alpha1"

		// External name: Linear-assigned UUID. The workspace settings
		// resource is a singleton; the provider assigns the identifier.
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

		// Sub-objects are managed as nested Terraform blocks. Upjet maps
		// them automatically:
		//   projects    → spec.forProvider.projects
		//     (updateReminderDay, updateReminderHour, updateReminderFrequency)
		//   initiatives → spec.forProvider.initiatives
		//     (enabled)
		//   feed        → spec.forProvider.feed
		//     (enabled, schedule)
		//   customers   → spec.forProvider.customers
		//     (enabled)
	})
}

// ValidateWorkspaceSettingsParams validates WorkspaceSettings resource
// parameters. Called by the reconciler before create and update operations.
func ValidateWorkspaceSettingsParams(params map[string]any) error {
	if v, ok := params["fiscal_year_start_month"]; ok {
		if month, ok := toInt(v); ok {
			if err := validation.ValidateFiscalMonth("fiscalYearStartMonth", month); err != nil {
				return err
			}
		}
	}
	return nil
}

// ErrSingletonViolation is returned when a second WorkspaceSettings resource
// is created for the same workspace.
var ErrSingletonViolation = fmt.Errorf("only one WorkspaceSettings resource is allowed per workspace (singleton)")

// CheckSingleton verifies that no other WorkspaceSettings resource exists for
// the workspace. The existingID is the external name of any existing
// WorkspaceSettings resource; currentID is the external name of the resource
// being reconciled. Returns ErrSingletonViolation if they differ (indicating
// a duplicate).
func CheckSingleton(existingID, currentID string) error {
	if existingID != "" && currentID != "" && existingID != currentID {
		return ErrSingletonViolation
	}
	return nil
}

// toInt converts a Terraform parameter value to int.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}
