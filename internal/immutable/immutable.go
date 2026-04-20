// Package immutable provides centralized immutability enforcement for
// Crossplane managed resources. It defines a generic mechanism to check
// that fields marked as immutable are not modified after resource creation.
package immutable

import "fmt"

// FieldConfig describes an immutable field on a resource.
type FieldConfig struct {
	// FieldName is the Terraform/spec field key (e.g., "team_id").
	FieldName string
	// DisplayName is the user-facing name shown in error messages (e.g., "teamId").
	DisplayName string
}

// ResourceConfig holds the immutability configuration for a resource type.
type ResourceConfig struct {
	// ResourceKind is the CRD kind (e.g., "TeamLabel", "WorkflowState").
	ResourceKind string
	// Fields lists all immutable fields for this resource.
	Fields []FieldConfig
}

// CheckImmutableField compares a single field between observed and desired
// state maps. Returns an error if the field value changed after creation.
// Empty observed values are skipped (resource not yet created).
func CheckImmutableField(observed, desired map[string]any, field FieldConfig) error {
	obs, _ := observed[field.FieldName].(string)
	des, _ := desired[field.FieldName].(string)
	if obs != "" && des != "" && obs != des {
		return fmt.Errorf(
			"%s is immutable after creation: cannot change from %q to %q",
			field.DisplayName, obs, des,
		)
	}
	return nil
}

// CheckImmutableFields validates all immutable fields for a resource
// configuration. Returns the first violation found, or nil if all fields
// are unchanged.
func CheckImmutableFields(observed, desired map[string]any, config ResourceConfig) error {
	for _, field := range config.Fields {
		if err := CheckImmutableField(observed, desired, field); err != nil {
			return err
		}
	}
	return nil
}

// --- Resource-specific immutability configurations ---

// TeamLabelConfig defines immutable fields for the TeamLabel resource.
var TeamLabelConfig = ResourceConfig{
	ResourceKind: "TeamLabel",
	Fields: []FieldConfig{
		{FieldName: "team_id", DisplayName: "teamId"},
	},
}

// WorkflowStateConfig defines immutable fields for the WorkflowState resource.
var WorkflowStateConfig = ResourceConfig{
	ResourceKind: "WorkflowState",
	Fields: []FieldConfig{
		{FieldName: "type", DisplayName: "type"},
		{FieldName: "team_id", DisplayName: "teamId"},
	},
}
