// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	releasepipeline "github.com/avodah-inc/provider-linear/internal/controller/linear/releasepipeline"
	releasestage "github.com/avodah-inc/provider-linear/internal/controller/linear/releasestage"
	team "github.com/avodah-inc/provider-linear/internal/controller/linear/team"
	teamlabel "github.com/avodah-inc/provider-linear/internal/controller/linear/teamlabel"
	teamworkflow "github.com/avodah-inc/provider-linear/internal/controller/linear/teamworkflow"
	template "github.com/avodah-inc/provider-linear/internal/controller/linear/template"
	user "github.com/avodah-inc/provider-linear/internal/controller/linear/user"
	workflowstate "github.com/avodah-inc/provider-linear/internal/controller/linear/workflowstate"
	workspacelabel "github.com/avodah-inc/provider-linear/internal/controller/linear/workspacelabel"
	workspacesettings "github.com/avodah-inc/provider-linear/internal/controller/linear/workspacesettings"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		releasepipeline.Setup,
		releasestage.Setup,
		team.Setup,
		teamlabel.Setup,
		teamworkflow.Setup,
		template.Setup,
		workflowstate.Setup,
		workspacelabel.Setup,
		workspacesettings.Setup,
		user.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		releasepipeline.SetupGated,
		releasestage.SetupGated,
		team.SetupGated,
		teamlabel.SetupGated,
		teamworkflow.SetupGated,
		template.SetupGated,
		workflowstate.SetupGated,
		workspacelabel.SetupGated,
		workspacesettings.SetupGated,
		user.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
