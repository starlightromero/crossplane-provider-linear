// Package team contains the Upjet resource configuration for linear_team.
package team

import (
	"fmt"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/avodah-inc/provider-linear/internal/validation"
)

// Configure adds custom resource configuration for linear_team.
//
// External name: uses the team `key` field (e.g., "ENG").
// Validations: key (team key pattern), name (min 2), color (hex),
// autoArchivePeriod (enum), autoClosePeriod (enum).
// Sub-objects: inline workflow states (backlog, unstarted, started,
// completed, canceled), triage, cycles, estimation.
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_team", func(r *ujconfig.Resource) {
		r.ShortGroup = "team"
		r.Kind = "Team"
		r.Version = "v1alpha1"

		// External name: the team key (e.g., "ENG") serves as the
		// Terraform import identifier.
		r.ExternalName = ujconfig.ExternalName{
			SetIdentifierArgumentFn: func(base map[string]any, externalName string) {
				base["key"] = externalName
			},
			GetExternalNameFn: func(tfstate map[string]any) (string, error) {
				if key, ok := tfstate["key"].(string); ok && key != "" {
					return key, nil
				}
				return "", fmt.Errorf("cannot determine external name: key field is empty")
			},
			OmittedFields: []string{"key"},
		}

		// Inline workflow state sub-objects are managed as nested
		// Terraform blocks. Upjet maps them automatically:
		//   backlog_workflow_state   → spec.forProvider.backlogWorkflowState
		//   unstarted_workflow_state → spec.forProvider.unstartedWorkflowState
		//   started_workflow_state   → spec.forProvider.startedWorkflowState
		//   completed_workflow_state → spec.forProvider.completedWorkflowState
		//   canceled_workflow_state  → spec.forProvider.canceledWorkflowState
		//
		// Triage, cycles, and estimation sub-objects are similarly
		// mapped as nested blocks by Upjet's schema bridge:
		//   triage     → spec.forProvider.triage
		//   cycles     → spec.forProvider.cycles
		//   estimation → spec.forProvider.estimation
	})
}

// ValidateTeamParams validates Team resource parameters. Called by the
// reconciler before create and update operations.
func ValidateTeamParams(params map[string]any) error {
	if err := validateTeamKey(params); err != nil {
		return err
	}
	if err := validateTeamNameParam(params); err != nil {
		return err
	}
	if err := validateTeamColor(params); err != nil {
		return err
	}
	if err := validateArchivePeriod(params); err != nil {
		return err
	}
	return validateClosePeriod(params)
}

func validateTeamKey(params map[string]any) error {
	key, ok := params["key"].(string)
	if !ok {
		return nil
	}
	return validation.ValidateTeamKey("key", key)
}

func validateTeamNameParam(params map[string]any) error {
	name, ok := params["name"].(string)
	if !ok {
		return nil
	}
	return validation.ValidateTeamName("name", name)
}

func validateTeamColor(params map[string]any) error {
	color, ok := params["color"].(string)
	if !ok || color == "" {
		return nil
	}
	return validation.ValidateHexColor("color", color)
}

func validateArchivePeriod(params map[string]any) error {
	v, ok := params["auto_archive_period"]
	if !ok {
		return nil
	}
	period, ok := toInt(v)
	if !ok {
		return nil
	}
	return validation.ValidateAutoArchivePeriod("autoArchivePeriod", period)
}

func validateClosePeriod(params map[string]any) error {
	v, ok := params["auto_close_period"]
	if !ok {
		return nil
	}
	period, ok := toInt(v)
	if !ok {
		return nil
	}
	return validation.ValidateAutoClosePeriod("autoClosePeriod", period)
}

// toInt converts a Terraform parameter value to int. Terraform may represent
// numbers as float64 (from JSON) or int.
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
