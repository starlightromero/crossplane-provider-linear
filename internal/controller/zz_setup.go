// Package controller contains the controller setup for provider-linear.
//
// This file is a placeholder that will be replaced by Upjet-generated code
// after running `make generate`. The generated version will register a
// reconciler for each managed resource type.
package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	tjcontroller "github.com/crossplane/upjet/v2/pkg/controller"
)

// Setup creates all Linear controllers with the supplied logger and adds them
// to the supplied manager.
//
// After Upjet code generation (`make generate`), this function will be
// replaced with generated code that registers controllers for:
//   - Team
//   - TeamLabel
//   - TeamWorkflow
//   - Template
//   - WorkflowState
//   - WorkspaceLabel
//   - WorkspaceSettings
//   - Workspace (data source)
func Setup(mgr ctrl.Manager, o tjcontroller.Options) error {
	return nil
}
