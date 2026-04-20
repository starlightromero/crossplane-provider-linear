// Package apis contains the API types for provider-linear.
//
// After Upjet code generation, this package will contain generated CRD types
// for all Linear managed resources and the Workspace data source.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToScheme adds all Linear resource types to the given scheme.
// This is a placeholder that will be replaced by Upjet-generated code
// after running `make generate`. The generated version will register
// CRD types for Team, TeamLabel, TeamWorkflow, Template, WorkflowState,
// WorkspaceLabel, WorkspaceSettings, and Workspace under the group
// linear.crossplane.io/v1alpha1.
func AddToScheme(s *runtime.Scheme) error {
	// After `make generate`, this will register:
	// - v1alpha1.Team
	// - v1alpha1.TeamLabel
	// - v1alpha1.TeamWorkflow
	// - v1alpha1.Template
	// - v1alpha1.WorkflowState
	// - v1alpha1.WorkspaceLabel
	// - v1alpha1.WorkspaceSettings
	// - v1alpha1.Workspace
	return nil
}
